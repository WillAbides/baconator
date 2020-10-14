package graph

import (
	"testing"
)

var globalPathLen int

func BenchmarkGraph_FindPath(b *testing.B) {
	var path []Node
	g := graphFromGob(b, "100k_graph.gob")
	b.Run("100k", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			src := Node(i % len(g.edgeIndex) - 2)
			dest := Node(i / 2 % len(g.edgeIndex) - 2)
			g.FindPath(&path, 999, src, dest, nil)
			globalPathLen = len(path)
		}
		b.ReportAllocs()
	})

	g = graphFromGob(b, "1MM_graph.gob")
	b.Run("1000k", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			src := Node(i % len(g.edgeIndex) - 2)
			dest := Node(i / 2 % len(g.edgeIndex) - 2)
			g.FindPath(&path, 999, src, dest, nil)
		}
		b.ReportAllocs()
		globalPathLen = len(path)
	})

	_ = globalPathLen
}
