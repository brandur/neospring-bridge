package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"

	"github.com/brandur/neospring-bridge/internal/util/stringutil"
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

// Should end with a slash.
const canonicalURL = "https://brandur.org/"

var (
	aHrefRE  = regexp.MustCompile(`<a href="/`)
	imgSrcRE = regexp.MustCompile(`<img src="/`)
)

func canonicalizeURLs(content string) string {
	content = aHrefRE.ReplaceAllString(content, `<a href="`+canonicalURL)
	content = imgSrcRE.ReplaceAllString(content, `<img src="`+canonicalURL)
	return content
}

func fetchFeed(ctx context.Context, url string) (*Feed, error) {
	data, err := requestWithRetries(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, xerrors.Errorf("error getting feed: %w", err)
	}

	var feed Feed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, xerrors.Errorf("error unmarshaling XML feed: %w", err)
	}

	return &feed, nil
}

var (
	// Photos are large and boards are expected to be displayed in small spaces, so
	// don't bother with the additional srcset information. It eats up quite a few
	// characters.
	srcSetRE = regexp.MustCompile(`\n? +srcset=".*?"`)

	twoPlusSpacesRE = regexp.MustCompile(` {2,}`)
)

func minimizeContent(content string) string {
	content = html.UnescapeString(content)
	content = srcSetRE.ReplaceAllString(content, "")
	content = strings.ReplaceAll(content, "\n", "")
	content = twoPlusSpacesRE.ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)
	return content
}

//go:embed layout.tmpl.html
var layout string

func renderLayout(title, content string, timestamp time.Time) (string, error) {
	tmpl, err := template.New("layout").Parse(layout)
	if err != nil {
		return "", xerrors.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, map[string]any{
		"Content":   content,
		"Timestamp": fmt.Sprintf(`<time datetime="%s">`, timestamp.Format(timestampFormat)),
		"Title":     title,
	}); err != nil {
		return "", xerrors.Errorf("error executing template: %w", err)
	}

	return buf.String(), nil
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

		for key, vals := range headers {
			for _, val := range vals {
				r.Header.Add(key, val)
			}
		}

		logger.Infof("Request: %s %v (attempt: %d)", method, url, requestNum-1)

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			outerErr = xerrors.Errorf("error making request: %w", err)
			continue
		}

		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			outerErr = xerrors.Errorf("error reading response body: %w", err)
			continue
		}

		logger.Infof("Response: %s (body: %q)", resp.Status, stringutil.SampleLong(string(respBody)))

		// Conflict is returned by a Spring '83 implementation in cases where a
		// newer version of a board has already been posted, so if we encounter
		// this, consider it a success and stop retrying.
		if resp.StatusCode >= 300 && resp.StatusCode != http.StatusConflict {
			err := xerrors.Errorf("bad status code during request: %d", resp.StatusCode)
			if shouldRetryStatusCode(resp.StatusCode) {
				outerErr = err
				continue
			}

			return nil, err
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

func updateSpring(ctx context.Context, keyPair *KeyPair, springURL string, entry *Entry) error {
	rendered, err := renderLayout(entry.Title, entry.Content.Content, entry.Published)
	if err != nil {
		return err
	}

	rendered = canonicalizeURLs(rendered)
	rendered = minimizeContent(rendered)

	logger.Infof("Raw content is %d bytes; %d bytes after layout, canonicalization, and minification",
		len([]byte(entry.Content.Content)),
		len(rendered),
	)

	respBody, err := requestWithRetries(ctx, http.MethodPut, springURL+"/"+keyPair.PublicKey, http.Header{
		"Spring-Signature": []string{keyPair.SignHex([]byte(rendered))},
	}, []byte(rendered))
	if err != nil {
		return xerrors.Errorf("error updating board: %w", err)
	}

	logger.Infof("Successfully published entry %q with timestamp %v (resp body: %q)",
		entry.Title,
		entry.Published,
		string(respBody),
	)

	return err
}
