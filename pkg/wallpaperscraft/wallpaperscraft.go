package wallpaperscraft

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

	//log.Println("Found the following images:")
	//log.Printf("%q\n", images)

	if len(images) > 0 {
		utils.CreateDirIfNotExist("./download/" + s.terms)
		semchan := make(chan struct{}, 3)
		for _, img := range images {
			semchan <- struct{}{}
			go func(img string) {
				if err := s.downloadImage(img); err != nil {
					log.Fatal(err)
				}
				<-semchan
			}(img)
		}
	}

	return nil
}

func (s *Scraper) getPage(pageNum int) (*goquery.Document, bool, error) {
	url := fmt.Sprintf("%s/search/?order=&page=%d&query=%s&size=", s.url, pageNum, s.terms)
	response, err := http.Get(url)
	if err != nil {
		return nil, false, err
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
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
		return nil
	}

	f, err := os.OpenFile("./download/"+s.terms+"/"+fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0744)
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