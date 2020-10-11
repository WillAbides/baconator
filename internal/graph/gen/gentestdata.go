// +build gen

package main

import (
	"flag"
	"log"

	"github.com/willabides/baconator/internal/graph"
)

func main() {
	var outputDir string
	flag.StringVar(&outputDir, "o", "", "output directory")
	flag.Parse()
	err := graph.GenerateTestData(outputDir)
	if err != nil {
		log.Fatal(err)
	}
}
