package main

import (
	"log"

	"github.com/roffe/bg-scraper/pkg/wallpapercave"
)

// Scraper our scraper interface
type Scraper interface {
	Scrape() error
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var pages = []Scraper{
		wallpapercave.New("warhammer"),
	}

	for _, p := range pages {
		if err := p.Scrape(); err != nil {
			panic(err)
		}
	}
}
