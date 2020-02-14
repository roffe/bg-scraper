package wallpaperscraft

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/roffe/bg-scraper/pkg/utils"
)

// Scraper ...
type Scraper struct {
	url, terms string
}

// New ...
func New(terms string) *Scraper {
	return &Scraper{
		url:   "https://wallpaperscraft.com",
		terms: terms,
	}
}

// Scrape ...
func (s *Scraper) Scrape() error {
	var images []string
	var err error
	var document *goquery.Document
	more := true
	page := 1
	for more {
		document, more, err = s.getPage(page)
		if err != nil {
			return err
		}
		document.Find("img").Each(func(index int, element *goquery.Selection) {
			if element.HasClass("wallpapers__image") {
				if src, exists := element.Attr("src"); exists {
					url := strings.ReplaceAll(src, "300x168", "1600x900")
					images = append(images, url)
				}
			}
		})
		page++
	}

	if len(images) > 0 {
		utils.CreateDirIfNotExist("./download/wallpaperscraft/" + s.terms)
		semchan := make(chan struct{}, 3)
		var wg sync.WaitGroup
		for _, img := range images {
			semchan <- struct{}{}
			wg.Add(1)
			go func(img string) {
				defer wg.Done()
				if err := s.downloadImage(img); err != nil {
					log.Fatal(err)
				}
				<-semchan
			}(img)
		}
		wg.Wait()
	}

	return nil
}

func (s *Scraper) getPage(pageNum int) (*goquery.Document, bool, error) {
	url := fmt.Sprintf("%s/search/?order=&page=%d&query=%s&size=", s.url, pageNum, s.terms)
	resp, err := http.Get(url)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("Error loading HTTP response body: %s", err)
	}

	var more = true
	document.Find("div").EachWithBreak(func(index int, element *goquery.Selection) bool {
		if element.HasClass("error") {
			more = false
			return false
		}
		return true
	})

	return document, more, nil
}

func (s *Scraper) downloadImage(img string) error {
	parts := strings.Split(img, "/")
	fname := parts[len(parts)-1]
	resp, err := http.Get(img)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("not 200, %s\n", img)
		return nil
	}

	f, err := os.OpenFile("./download/wallpaperscraft/"+s.terms+"/"+fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	log.Println("Downloaded ", written, "to ", f.Name())
	return nil
}
