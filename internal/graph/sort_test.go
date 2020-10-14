package graph

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/require"
)

func Test_sortNodesBYOB(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		test := []NodeIdx{3, 1000, 1, 100, 0, 999, math.MaxUint32}
		b := make([]NodeIdx, len(test))
		sortNodesBYOB(test, b)
		require.True(t, nodesAreSorted(test))
	})

	t.Run("empty", func(t *testing.T) {
		test := []NodeIdx{}
		b := []NodeIdx{}
		sortNodesBYOB(test, b)
		require.Empty(t, test)
	})

	t.Run("rand", func(t *testing.T) {
		test := func(data []NodeIdx) bool {
			buffer := make([]NodeIdx, len(data))
			sortNodesBYOB(data, buffer)
			return nodesAreSorted(data)
		}
		if err := quick.Check(test, nil); err != nil {
			t.Error(err)
		}
	})
}

func nodesAreSorted(data []NodeIdx) bool {
	for idx, x := range data {
		if idx == 0 {
			continue
		}
		if x < data[idx-1] {
			return false
		}
	}
	return true
}
