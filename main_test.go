package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldRetryStatusCode(t *testing.T) {
	require.True(t, shouldRetryStatusCode(http.StatusTooManyRequests))
	require.True(t, shouldRetryStatusCode(http.StatusInternalServerError))

	// Conflict is returned by a Spring '83 implementation in cases where a
	// newer version of a board has already been posted, so if we encounter
	// this, consider it a success and stop retrying.
	require.False(t, shouldRetryStatusCode(http.StatusConflict))
}
