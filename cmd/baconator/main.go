package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/willabides/baconator"
)

func main() {
	var datafile string
	var tcpAddr string
	flag.StringVar(&datafile, "data", "data.txt.bz2", "path to data.txt.bz2")
	flag.StringVar(&tcpAddr, "l", "localhost:8239", "tcp address to listen on")
	flag.Parse()
	b := &baconator.Baconator{}
	log.Printf("loading data from %s", datafile)
	err := b.LoadFromDatafile(datafile)
	if err != nil {
		log.Fatalf("error loading data: %v", err)
	}
	s := baconator.NewServer(b)
	log.Printf("Listening at %s", tcpAddr)
	err = http.ListenAndServe(tcpAddr, s)
	if err != nil {
		log.Fatal(err)
	}
}
