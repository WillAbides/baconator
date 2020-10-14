// +build gen

package graph

import (
	"encoding/gob"
	"math/rand"
	"os"
	"path/filepath"
)

type testDataOpts struct {
	randSource   rand.Source
	nodeCount    int
	minNeighbors int
	maxNeighbors int
}

// GenerateTestData generates test data
func GenerateTestData(outputDir string) error {
	err := os.MkdirAll(outputDir, 0o700)
	if err != nil {
		return err
	}
	for filename, opts := range map[string]*testDataOpts{
		"100k_graph.gob": {
			nodeCount:    100_000,
			minNeighbors: 1,
			maxNeighbors: 15,
		},
		"1MM_graph.gob": {
			nodeCount:    1_000_000,
			minNeighbors: 1,
			maxNeighbors: 15,
		},
	} {
		filename = filepath.Join(outputDir, filename)
		err = generateTestDataFile(filename, false, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateTestDataFile(filename string, force bool, opts *testDataOpts) error {
	exists := true
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		exists = false
	} else if err != nil {
		return err
	}
	if exists && !force {
		return nil
	}
	data := genTestData(opts)
	graph := New(data)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	return gob.NewEncoder(file).Encode(graph)
}

func genTestData(opts *testDataOpts) [][]NodeIdx {
	if opts == nil {
		opts = &testDataOpts{}
	}
	randSource := opts.randSource
	if randSource == nil {
		randSource = rand.NewSource(1)
	}
	nodeCount := opts.nodeCount
	if nodeCount == 0 {
		nodeCount = 100
	}
	minNeighbors := opts.minNeighbors
	maxNeighbors := opts.maxNeighbors
	if maxNeighbors == 0 {
		maxNeighbors = 10
	}
	sliceSize := 2 * maxNeighbors
	bigSlice := make([]NodeIdx, nodeCount*(int(float64(sliceSize)*1.1)))
	rnd := rand.New(randSource)
	data := make([][]NodeIdx, nodeCount)
	for i := 0; i < nodeCount; i++ {
		if len(data[i]) == 0 {
			allocateNodes(sliceSize, &data[i], &bigSlice)
		}
		neighborCount := rnd.Intn(maxNeighbors-minNeighbors) + minNeighbors
		for j := 0; j < neighborCount; j++ {
			neighbor := NodeIdx(rnd.Intn(nodeCount))
			if neighbor == NodeIdx(i) {
				continue
			}
			data[i] = append(data[i], neighbor)
			if len(data[neighbor]) == 0 {
				allocateNodes(sliceSize, &data[neighbor], &bigSlice)
			}
			data[neighbor] = append(data[neighbor], NodeIdx(i))
		}
	}
	return data
}

func allocateNodes(size int, nodes, buf *[]NodeIdx) {
	if len(*nodes) > 0 {
		return
	}
	if len(*buf) >= size {
		*nodes = (*buf)[0:0:size]
		*buf = (*buf)[size:]
	} else {
		*nodes = make([]NodeIdx, 0, size)
	}
}
