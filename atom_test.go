package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestSortEntriesDesc(t *testing.T) {
	now := time.Now()

	e1 := &Entry{Published: now.Add(1 * time.Second)}
	e2 := &Entry{Published: now.Add(2 * time.Second)}
	e3 := &Entry{Published: now.Add(3 * time.Second)}

	es := []*Entry{e2, e1, e3}
	slices.SortFunc(es, sortEntriesDesc)

	require.Equal(t, []*Entry{e3, e2, e1}, es)
}
