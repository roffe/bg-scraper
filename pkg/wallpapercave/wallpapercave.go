package wallpapercave

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/roffe/bg-scraper/pkg/utils"
)

// Scraper ...
type Scraper struct {
	URL        url.URL
	url, terms string
}

// New returns a new wallpapercave scraper
func New(terms string) *Scraper {
	return &Scraper{
		url:   "https://wallpapercave.com",
		terms: terms,
	}
}

// Scrape ...
func (s *Scraper) Scrape() error {
	searchURL := s.url + "/search?q=" + s.terms
	response, err := http.Get(searchURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return fmt.Errorf("Error loading HTTP response body: %s", err)
	}

	var albums []string
	document.Find("a").Each(func(index int, element *goquery.Selection) {
		if element.HasClass("albumthumbnail") {
			if href, exists := element.Attr("href"); exists {
				albums = append(albums, href)
			}
		}
	})

	//log.Println("Found the following albums:")
	//log.Printf("%q\n", albums)

	var toDownload []image

	for _, album := range albums {
		res := s.parseAlbum(album)
		toDownload = append(toDownload, res...)
	}
	semchan := make(chan struct{}, 3)
	var wg sync.WaitGroup
	for _, img := range toDownload {
		semchan <- struct{}{}
		wg.Add(1)
		go func(img image) {
			defer wg.Done()
			if err := s.downloadImage(img); err != nil {
				log.Fatal(err)
			}
			<-semchan
		}(img)
	}
	wg.Wait()

	return nil
}

type image struct {
	id, slug string
}

func (s *Scraper) parseAlbum(album string) (images []image) {
	utils.CreateDirIfNotExist("./download" + album)

	response, err := http.Get(s.url + album)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	document.Find("img").Each(func(index int, element *goquery.Selection) {
		if element.HasClass("wpimg") {
			id, exists := element.Attr("data-url")
			if !exists {
				return
			}
			slug, exists := element.Attr("slug")
			if !exists {
				return
			}
			images = append(images, image{id: id, slug: slug})

		}

	})

	return images
}

func (s *Scraper) downloadImage(img image) error {
	downloadURL := fmt.Sprintf("%s/download/%s-%s", s.url, img.slug, img.id)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, params, err := mime.ParseMediaType(resp.Header.Get("content-disposition"))
	if err != nil {
		return err
	}
	filename, ok := params["filename"]
	if !ok {
		filename = img.id + ".jpg"
	}

	f, err := os.OpenFile("./download/"+img.slug+"/"+filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	log.Println("Downloaded", written, "bytes to", f.Name())
	return nil
}
