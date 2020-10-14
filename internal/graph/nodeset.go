package graph

const (
	nodeSetBucketBits = 5
	nodeSetBucketSize = 1 << 5
	nodeSetBucketMask = nodeSetBucketSize - 1
)

type parentsMap struct {
	parents []NodeIdx
	nodeSet []uint32
}

func newParentsMap(size int) *parentsMap {
	return &parentsMap{
		nodeSet: make([]uint32, (size+31)/32),
		parents: make([]NodeIdx, size),
	}
}

func (p *parentsMap) clear() {
	for i := range p.nodeSet {
		p.nodeSet[i] = 0
	}
}

func (p *parentsMap) contains(node NodeIdx) bool {
	bucket := uint32(node >> nodeSetBucketBits)
	bit := uint32(1 << (node & nodeSetBucketMask))
	return p.nodeSet[bucket]&bit != 0
}

func (p *parentsMap) setParent(node, parent NodeIdx) {
	bucket := uint32(node >> nodeSetBucketBits)
	bit := uint32(1 << (node & nodeSetBucketMask))
	p.nodeSet[bucket] |= bit

	p.parents[node] = parent
}

func (p *parentsMap) getParent(node NodeIdx) NodeIdx {
	return p.parents[node]
}
