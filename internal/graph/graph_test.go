package graph

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate go run -tags gen ./gen -o ./testdata

func graphFromGob(t testing.TB, filename string) *Graph {
	t.Helper()
	filename = filepath.Join("testdata", filename)
	if !assert.FileExists(t, filename, "file doesn't exist. try running script/generate") {
		return nil
	}
	file, err := os.Open(filename)
	if !assert.NoError(t, err) {
		return nil
	}
	var graph Graph
	err = gob.NewDecoder(file).Decode(&graph)
	if !assert.NoError(t, err) {
		return nil
	}
	return &graph
}

func TestNew(t *testing.T) {
	t.Run("neighborhood", func(t *testing.T) {
		neighbors := [][]Node{
			0: {1},
			1: {0, 2},
			2: {1, 3},
			3: {1, 2, 4},
			4: {3, 5},
			5: {4, 6},
			6: {5, 7},
			7: {6},
		}
		g := New(neighbors)
		require.Equal(t, 9, len(g.edgeIndex))
	})
}

func TestGraph_FindPath(t *testing.T) {
	t.Run("", func(t *testing.T) {
		neighbors := [][]Node{
			0: {1},
			1: {0, 2},
			2: {1, 3},
			3: {1, 2, 4},
			4: {3, 5},
			5: {4, 6},
			6: {5, 7},
			7: {6},
		}
		g := New(neighbors)

		tst := func(tt *testing.T, src, dest Node, want []Node) {
			tt.Helper()
			path := []Node{}
			g.FindPath(&path, 0, src, dest, nil)
			assert.Equal(tt, want, path)
		}

		tst(t, 0, 1, []Node{0, 1})
		tst(t, 0, 2, []Node{0, 1, 2})
		tst(t, 1, 2, []Node{1, 2})
		tst(t, 1, 3, []Node{1, 2, 3})
		tst(t, 1, 4, []Node{1, 2, 3, 4})
		tst(t, 1, 5, []Node{1, 2, 3, 4, 5})
		tst(t, 1, 6, []Node{1, 2, 3, 4, 5, 6})
		tst(t, 1, 7, []Node{1, 2, 3, 4, 5, 6, 7})
		tst(t, 1, 8, []Node{})
		tst(t, 1, 99, []Node{})
	})

	t.Run("priority", func(t *testing.T) {
		neighbors := [][]Node{
			0: {1},
			1: {0, 2},
			2: {1, 3, 4},
			3: {2, 5},
			4: {2, 5},
			5: {3, 4, 6},
			6: {5, 7},
			7: {6},
		}
		g := New(neighbors)
		var path []Node
		g.FindPath(&path, 0, 0, 7, func(node Node) int64 {
			if node == 4 {
				return 1
			}
			return 0
		})
		require.Equal(t, []Node{0, 1, 2, 4, 5, 6, 7}, path)
	})
}

func TestGraph_NodeNeighbors(t *testing.T) {
	neighbors := [][]Node{
		0: {1},
		1: {2, 0},
		2: {1, 3},
		3: {1, 2, 4},
		4: {3, 5},
		5: {4},
	}
	g := New(neighbors)
	for n, neighbors := range neighbors {
		require.Equal(t, neighbors, g.NodeNeighbors(Node(n)))
	}
}
