package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"
)

// From spec: <time datetime="YYYY-MM-DDTHH:MM:SSZ">.
const timestampFormat = "2006-01-02T15:04:05Z"

var logger = logrus.New()

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		abortErr(err)
	}
}

func abort(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func abortErr(err error) {
	abort("error: %v", err)
}

func fetchFeed(ctx context.Context, url string) (*Feed, error) {
	data, err := requestWithRetries(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, err
	}

	var feed Feed
	if xml.Unmarshal(data, &feed); err != nil {
		return nil, xerrors.Errorf("error unmarshaling XML feed: %w", err)
	}

	return &feed, nil
}

func requestWithRetries(ctx context.Context, method, url string, headers http.Header, body []byte) ([]byte, error) {
	var outerErr error
	var requestNum int

	for {
		switch {
		case requestNum > 2:
			return nil, outerErr
		case requestNum > 0:
			time.Sleep(time.Duration(math.Pow(2, float64(requestNum))) * time.Second)
		}
		requestNum++

		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		r, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, xerrors.Errorf("error creating new request: %w", err)
		}

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			outerErr = xerrors.Errorf("error making request: %w", err)
			continue
		}

		// Conflict is returned by a Spring '83 implementation in cases where a
		// newer version of a board has already been posted, so if we encounter
		// this, consider it a success and stop retrying.
		if resp.StatusCode >= 300 && resp.StatusCode != http.StatusConflict {
			if shouldRetryStatusCode(resp.StatusCode) {
				outerErr = xerrors.Errorf("bad status code during request: %d", resp.StatusCode)
				continue
			}
		}

		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			outerErr = xerrors.Errorf("error reading response body: %w", err)
			continue
		}

		return respBody, nil
	}
}

func run(ctx context.Context) error {
	type Config struct {
		AtomFeedURL      string `env:"ATOM_FEED_URL,required"`
		SpringPrivateKey string `env:"SPRING_PRIVATE_KEY,required"`
		SpringPublicKey  string `env:"SPRING_PUBLIC_KEY,required"`
		SpringURL        string `env:"SPRING_URL,required"`
	}

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return xerrors.Errorf("error parsing env config: %w", err)
	}

	keyPair, err := ParseKeyPairUnchecked(config.SpringPrivateKey)
	if err != nil {
		return err
	}

	// Not strictly needed, but just make sure that one isn't accidentally
	// updated without the other.
	if config.SpringPublicKey != keyPair.PublicKey {
		return xerrors.Errorf("SPRING_PUBLIC_KEY doesn't match the public key portion of SPRING_PRIVATE_KEY")
	}

	feed, err := fetchFeed(ctx, config.AtomFeedURL)
	if err != nil {
		return err
	}

	if len(feed.Entries) < 1 {
		logger.Infof("No entries in feed; taking no action")
		return nil
	}

	slices.SortFunc(feed.Entries, sortEntriesDesc)

	if err := updateSpring(ctx, keyPair, config.SpringURL, feed.Entries[0]); err != nil {
		return err
	}

	return nil
}

func shouldRetryStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests:
		fallthrough
	case http.StatusInternalServerError:
		return true
	}

	return false
}

func updateSpring(ctx context.Context, keyPair *KeyPair, url string, entry *Entry) error {
	content := fmt.Sprintf(`<time datetime="%s">\n\n%s\n\n%s`,
		entry.Published.Format(timestampFormat),
		entry.Title,
		entry.Content.Content,
	)

	_, err := requestWithRetries(ctx, http.MethodPut, url, http.Header{
		"Spring-Signature": []string{keyPair.SignHex([]byte(content))},
	}, []byte(content))
	return err
}
