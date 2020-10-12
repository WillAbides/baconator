package graph

import (
	"math/rand"
	"testing"
)

func BenchmarkGraph_FindPath(b *testing.B) {
	var path []Node
	rnd := rand.New(rand.NewSource(0))
	b.Run("100k", func(b *testing.B) {
		g := graphFromGob(b, "100k_graph.gob")
		llen := len(g.edgeIndex) - 1
		src := Node(rnd.Intn(llen))
		dest := Node(rnd.Intn(llen))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			g.FindPath(&path, 999, src, dest, nil)
		}
		b.ReportAllocs()
	})

	b.Run("1000k", func(b *testing.B) {
		g := graphFromGob(b, "1MM_graph.gob")
		llen := len(g.edgeIndex) - 1
		src := Node(rnd.Intn(llen))
		dest := Node(rnd.Intn(llen))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			g.FindPath(&path, 999, src, dest, nil)
		}
		b.ReportAllocs()
	})
}
