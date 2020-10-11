package graph

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// Node is a graph node
type Node uint32

// Graph is a graph of connected Nodes
type Graph struct {
	neighborIndex  []int
	neighbors      []Node
	slicePool      sync.Pool
	parentsMapPool sync.Pool
}

// New creates a new Graph
//  NodeNeighbors is a list of neighbors where the index is the node id and the
//   value is that node's neighbors
//  NodeNeighbors[0] is special. It isn't a valid node in the Graph. It can have no
//   neighbors and no node can have zero as a neighbor.
func New(nodeNeighbors [][]Node) (*Graph, error) {
	if len(nodeNeighbors) == 0 {
		return nil, fmt.Errorf("nodeNeighbors can't be empty")
	}
	var neighborSize int
	for _, neighbors := range nodeNeighbors {
		neighborSize += len(neighbors)
	}
	var g Graph
	g.neighbors = make([]Node, 0, neighborSize)
	g.neighborIndex = make([]int, 0, len(nodeNeighbors)+1)
	g.neighborIndex = append(g.neighborIndex, len(g.neighbors))
	for _, neighbors := range nodeNeighbors {
		g.neighbors = append(g.neighbors, neighbors...)
		g.neighborIndex = append(g.neighborIndex, len(g.neighbors))
	}
	g.createPools()
	return &g, nil
}

// Neighbors returns a copy the neighbors slice. This is memory intensive,
//  so don't use it in production code.
func (g *Graph) Neighbors() []Node {
	res := make([]Node, len(g.neighbors))
	copy(res, g.neighbors)
	return res
}

// NeighborIndex returns a copy of the neighborIndex slice.  This is memory intensive,
//  so don't use it in production code.
func (g *Graph) NeighborIndex() []int {
	res := make([]int, len(g.neighborIndex))
	copy(res, g.neighborIndex)
	return res
}

func (g *Graph) createPools() {
	g.slicePool = sync.Pool{
		New: func() interface{} {
			slice := make([]Node, 0, len(g.neighbors))
			return &slice
		},
	}
	g.parentsMapPool = sync.Pool{
		New: func() interface{} {
			return newParentsMap(len(g.neighborIndex) - 1)
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

func (g *Graph) borrowLevelSlice() *[]Node {
	s := g.slicePool.Get().(*[]Node)
	*s = (*s)[:0]
	return s
}

func (g *Graph) returnLevelSlice(slice *[]Node) {
	g.slicePool.Put(slice)
}

// NodeNeighbors returns n's immediate neighbors
func (g *Graph) NodeNeighbors(n Node) []Node {
	start, end := g.neighborIndex[n], g.neighborIndex[n+1]
	return g.neighbors[start:end]
}

// FindLevels returns the hop count of each node in the graph
func (g *Graph) FindLevels(source Node) []int {
	size := len(g.neighborIndex) - 1
	level := make([]int, size)
	currentLevel := make([]Node, 0, size)
	nextLevel := make([]Node, 0, size)
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
type PriorityFunc func(node Node) int64

// FindPath finds the shortest path from source to dest
//  When there are multiple shortest paths, FindPath may return any one of those paths.
//  The first element of the returned path is always source, and the last element is always dest.
//  path - is a pointer to a slice that FindPath will set to the found path
//  When no path is found, path will be set to zero length
func (g *Graph) FindPath(path *[]Node, maxPathLength int, source, dest Node, priorityFn PriorityFunc) {
	const defaultMaxPathLength = 9
	if maxPathLength <= 0 {
		maxPathLength = defaultMaxPathLength
	}
	size := len(g.neighborIndex) - 1
	if source >= Node(size) || dest >= Node(size) {
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
	var midPoint Node
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

func setPathLen(p *[]Node, length int) {
	if cap(*p) >= length {
		*p = (*p)[:length]
		return
	}
	*p = (*p)[:cap(*p)]
	extra := length - len(*p)
	*p = append(*p, make([]Node, extra)...)
}

func (g *Graph) nextLevel(currentLevel, scratchBuffer *[]Node, parents, otherParents *parentsMap, priority PriorityFunc) (Node, bool) {
	*scratchBuffer = (*scratchBuffer)[:0]
	var midPoint Node
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

func prioritySort(nodes *[]Node, fn PriorityFunc) {
	sort.Slice(*nodes, func(i, j int) bool {
		return fn((*nodes)[i]) > fn((*nodes)[j])
	})
}

type graphSerializer struct {
	Neighbors     []Node `json:"neighbors"`
	NeighborIndex []int  `json:"neighborIndex"`
}

// GobDecode implements gob.GobDecoder
func (g *Graph) GobDecode(p []byte) error {
	var gs graphSerializer
	err := gob.NewDecoder(bytes.NewReader(p)).Decode(&gs)
	if err != nil {
		return err
	}
	g.neighbors = gs.Neighbors
	g.neighborIndex = gs.NeighborIndex
	g.createPools()
	return nil
}

// GobEncode implements gob.GobEncoder
func (g *Graph) GobEncode() ([]byte, error) {
	gs := graphSerializer{
		Neighbors:     g.neighbors,
		NeighborIndex: g.neighborIndex,
	}
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&gs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (g *Graph) UnmarshalJSON(p []byte) error {
	var gs graphSerializer
	err := json.Unmarshal(p, &gs)
	if err != nil {
		return err
	}
	g.neighbors = gs.Neighbors
	g.neighborIndex = gs.NeighborIndex
	g.createPools()
	return nil
}

// MarshalJSON implements json.Marshaler
func (g *Graph) MarshalJSON() ([]byte, error) {
	gs := graphSerializer{
		Neighbors:     g.neighbors,
		NeighborIndex: g.neighborIndex,
	}
	return json.Marshal(&gs)
}
