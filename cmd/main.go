package main

import (
	"log"
	"os"

	bgscraper "github.com/roffe/bg-scraper"
	"github.com/roffe/bg-scraper/pkg/wallpaperscraft"
)

var searchArgs string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) != 2 {
		log.Fatal("Missing search parameter")
	}
	searchArgs = os.Args[1]
}

func main() {
	var pages = []bgscraper.Scraper{
		//wallpapercave.New(searchArgs),
		wallpaperscraft.New(searchArgs),
	}

	for _, p := range pages {
		if err := p.Scrape(); err != nil {
			log.Fatal(err)
		}
	}
}
