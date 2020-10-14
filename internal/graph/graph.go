package graph

import (
	"bytes"
	"encoding/gob"
	"sort"
	"sync"
)

// NodeIdx is the index where a node can be found
type NodeIdx uint32

// Graph is a graph of connected Nodes
type Graph struct {
	// edgeIndex is used to find edges for a source node.
	//
	// The key is the source node and the value is the position in edgeTargets
	// where the slice of the node's edge targets begins.
	edgeIndex []int

	edgeTargets    []NodeIdx
	slicePool      sync.Pool
	parentsMapPool sync.Pool
}

// New creates a new Graph
//
// nodeNeighbors is a list of neighbors where the index is the node id and the
// value is that node's neighbors
func New(nodeNeighbors [][]NodeIdx) *Graph {
	var neighborSize int
	for _, neighbors := range nodeNeighbors {
		neighborSize += len(neighbors)
	}
	var g Graph
	g.edgeTargets = make([]NodeIdx, 0, neighborSize)
	g.edgeIndex = make([]int, 0, len(nodeNeighbors)+1)
	g.edgeIndex = append(g.edgeIndex, len(g.edgeTargets))
	for _, neighbors := range nodeNeighbors {
		g.edgeTargets = append(g.edgeTargets, neighbors...)
		g.edgeIndex = append(g.edgeIndex, len(g.edgeTargets))
	}
	g.createPools()
	return &g
}

func (g *Graph) createPools() {
	g.slicePool = sync.Pool{
		New: func() interface{} {
			slice := make([]NodeIdx, 0, len(g.edgeTargets))
			return &slice
		},
	}
	g.parentsMapPool = sync.Pool{
		New: func() interface{} {
			return newParentsMap(len(g.edgeIndex) - 1)
		},
	}
}

func (g *Graph) borrowParentsMap() *parentsMap {
	return g.parentsMapPool.Get().(*parentsMap)
}

func (g *Graph) returnParentsMap(mp *parentsMap) {
	mp.clear()
	g.parentsMapPool.Put(mp)
}

func (g *Graph) borrowLevelSlice() *[]NodeIdx {
	s := g.slicePool.Get().(*[]NodeIdx)
	*s = (*s)[:0]
	return s
}

func (g *Graph) returnLevelSlice(slice *[]NodeIdx) {
	g.slicePool.Put(slice)
}

// NodeNeighbors returns n's immediate neighbors
func (g *Graph) NodeNeighbors(n NodeIdx) []NodeIdx {
	start, end := g.edgeIndex[n], g.edgeIndex[n+1]
	return g.edgeTargets[start:end]
}

// FindLevels returns the hop count of each node in the graph
func (g *Graph) FindLevels(source NodeIdx) []int {
	size := len(g.edgeIndex) - 1
	level := make([]int, size)
	currentLevel := make([]NodeIdx, 0, size)
	nextLevel := make([]NodeIdx, 0, size)
	visited := newParentsMap(size)
	level[source] = 1
	currentLevel = append(currentLevel, source)
	visited.setParent(source, 0)

	levelNumber := 2
	for len(currentLevel) > 0 {
		for _, node := range currentLevel {
			for _, neighbor := range g.NodeNeighbors(node) {
				if !visited.contains(neighbor) {
					visited.setParent(neighbor, 0)
					nextLevel = append(nextLevel, neighbor)
				}
			}
		}
		sortNodesBYOB(nextLevel, currentLevel[:cap(currentLevel)])
		for _, neighbor := range nextLevel {
			level[neighbor] = levelNumber
		}
		levelNumber++
		currentLevel = currentLevel[:0:cap(currentLevel)]
		currentLevel, nextLevel = nextLevel, currentLevel
	}
	return level
}

// PriorityFunc returns a node's priority when choosing between nodes.  This is not cost.
//  The shortest path still wins no matter the priority. Higher number is higher priority. 2 gets
//  chosen before 1.
type PriorityFunc func(node NodeIdx) int64

// FindPath finds the shortest path from source to dest
//  When there are multiple shortest paths, FindPath may return any one of those paths.
//  The first element of the returned path is always source, and the last element is always dest.
//  path - is a pointer to a slice that FindPath will set to the found path
//  When no path is found, path will be set to zero length
func (g *Graph) FindPath(path *[]NodeIdx, maxPathLength int, source, dest NodeIdx, priorityFn PriorityFunc) {
	const defaultMaxPathLength = 9
	if maxPathLength <= 0 {
		maxPathLength = defaultMaxPathLength
	}
	size := len(g.edgeIndex) - 1
	if source >= NodeIdx(size) || dest >= NodeIdx(size) {
		setPathLen(path, 0)
		return
	}

	if source == dest {
		setPathLen(path, 1)
		(*path)[0] = source
		return
	}

	srcCurrentLevel := g.borrowLevelSlice()
	defer g.returnLevelSlice(srcCurrentLevel)
	destCurrentLevel := g.borrowLevelSlice()
	defer g.returnLevelSlice(destCurrentLevel)
	scratchBuffer := g.borrowLevelSlice()
	defer g.returnLevelSlice(scratchBuffer)

	srcParentsMap := g.borrowParentsMap()
	defer g.returnParentsMap(srcParentsMap)
	destParentsMap := g.borrowParentsMap()
	defer g.returnParentsMap(destParentsMap)

	*srcCurrentLevel = append(*srcCurrentLevel, source)
	*destCurrentLevel = append(*destCurrentLevel, dest)
	var midPoint NodeIdx
	midFound := false
	srcPathLen := 1
	destPathLen := 1
	midFoundBySource := false
	for len(*srcCurrentLevel) > 0 && len(*destCurrentLevel) > 0 {
		midPoint, midFound = g.nextLevel(srcCurrentLevel, scratchBuffer, srcParentsMap, destParentsMap, priorityFn)
		if midFound || srcPathLen+destPathLen >= maxPathLength {
			midFoundBySource = true
			break
		}
		srcPathLen++
		midPoint, midFound = g.nextLevel(destCurrentLevel, scratchBuffer, destParentsMap, srcParentsMap, priorityFn)

		if midFound || srcPathLen+destPathLen >= maxPathLength {
			break
		}
		destPathLen++
	}
	if !midFound {
		*path = (*path)[:0]
		return
	}
	if midPoint == source {
		setPathLen(path, 2)
		(*path)[0] = source
		(*path)[1] = dest
		return
	}

	setPathLen(path, srcPathLen+destPathLen)
	(*path)[0] = source
	(*path)[len(*path)-1] = dest
	pathIdx := 0
	for n := srcParentsMap.getParent(midPoint); n != source; n = srcParentsMap.getParent(n) {
		pathIdx++
		idx := srcPathLen - pathIdx
		if !midFoundBySource {
			idx--
		}
		(*path)[idx] = n
	}
	pathIdx++
	(*path)[pathIdx] = midPoint
	for n := destParentsMap.getParent(midPoint); n != dest; n = destParentsMap.getParent(n) {
		pathIdx++
		(*path)[pathIdx] = n
	}
}

func setPathLen(p *[]NodeIdx, length int) {
	if cap(*p) >= length {
		*p = (*p)[:length]
		return
	}
	*p = (*p)[:cap(*p)]
	extra := length - len(*p)
	*p = append(*p, make([]NodeIdx, extra)...)
}

func (g *Graph) nextLevel(currentLevel, scratchBuffer *[]NodeIdx, parents, otherParents *parentsMap, priority PriorityFunc) (NodeIdx, bool) {
	*scratchBuffer = (*scratchBuffer)[:0]
	var midPoint NodeIdx
	foundMid := false
	levelLen := len(*currentLevel)
	for i := 0; i < levelLen && !foundMid; i++ {
		node := (*currentLevel)[i]
		neighbors := g.NodeNeighbors(node)
		if priority != nil {
			prioritySort(&neighbors, priority)
		}
		nLen := len(neighbors)
		for j := 0; j < nLen && !foundMid; j++ {
			neighbor := neighbors[j]
			if !parents.contains(neighbor) {
				parents.setParent(neighbor, node)
				*scratchBuffer = append(*scratchBuffer, neighbor)
			}
			if otherParents.contains(neighbor) {
				foundMid = true
				midPoint = neighbor
			}
		}
	}
	*currentLevel, *scratchBuffer = *scratchBuffer, *currentLevel
	return midPoint, foundMid
}

func prioritySort(nodes *[]NodeIdx, fn PriorityFunc) {
	sort.Slice(*nodes, func(i, j int) bool {
		return fn((*nodes)[i]) > fn((*nodes)[j])
	})
}

type graphSerializer struct {
	EdgeTargets []NodeIdx `json:"neighbors"`
	EdgeIndex   []int     `json:"edgeIndex"`
}

// GobDecode implements gob.GobDecoder
func (g *Graph) GobDecode(p []byte) error {
	var gs graphSerializer
	err := gob.NewDecoder(bytes.NewReader(p)).Decode(&gs)
	if err != nil {
		return err
	}
	g.edgeTargets = gs.EdgeTargets
	g.edgeIndex = gs.EdgeIndex
	g.createPools()
	return nil
}

// GobEncode implements gob.GobEncoder
func (g *Graph) GobEncode() ([]byte, error) {
	gs := graphSerializer{
		EdgeTargets: g.edgeTargets,
		EdgeIndex:   g.edgeIndex,
	}
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&gs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
