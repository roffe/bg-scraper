package main

import (
	"log"
	"os"

	"github.com/roffe/bg-scraper/pkg/wallpapercave"
)

// Scraper our scraper interface
type Scraper interface {
	Scrape() error
}

var searchArgs string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) != 2 {
		log.Fatal("Missing search parameter")
	}
	searchArgs = os.Args[1]
}

func main() {
	var pages = []Scraper{
		wallpapercave.New(searchArgs),
	}

	for _, p := range pages {
		if err := p.Scrape(); err != nil {
			panic(err)
		}
	}
}
