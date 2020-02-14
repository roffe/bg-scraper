package main

import (
	"log"
	"os"
	"sync"

	bgscraper "github.com/roffe/bg-scraper"
	"github.com/roffe/bg-scraper/pkg/desktopnexus"
	"github.com/roffe/bg-scraper/pkg/wallpapercave"
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
		desktopnexus.New(searchArgs),
		wallpapercave.New(searchArgs),
		wallpaperscraft.New(searchArgs),
	}

	var wg sync.WaitGroup
	semchan := make(chan struct{}, 3)
	for _, p := range pages {
		semchan <- struct{}{}
		wg.Add(1)
		go func(p bgscraper.Scraper) {
			defer wg.Done()
			if err := p.Scrape(); err != nil {
				log.Fatal(err)
			}
			<-semchan
		}(p)
	}
	wg.Wait()
	log.Println("Done!")
}
