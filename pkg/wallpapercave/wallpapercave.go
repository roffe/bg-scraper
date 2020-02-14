package wallpapercave

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Scraper ...
type Scraper struct {
	url, terms, agent string
}

// New returns a new wallpapercave scraper
func New(terms string) *Scraper {
	return &Scraper{
		url:   "https://wallpapercave.com",
		terms: terms,
		agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.87 Safari/537.36 Edg/80.0.361.50",
	}
}

// Scrape ...
func (w *Scraper) Scrape() error {
	searchURL := w.url + "/search?q=" + w.terms
	response, err := http.Get(searchURL)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	// Find all links and process them with the function
	// defined earlier
	var albums []string
	document.Find("a").Each(func(index int, element *goquery.Selection) {
		if element.HasClass("albumthumbnail") {
			if href, exists := element.Attr("href"); exists {
				albums = append(albums, href)
			}
		}
	})

	fmt.Printf("%q\n", albums)

	var toDownload []image

	for _, album := range albums {
		res := w.parseAlbum(album)
		toDownload = append(toDownload, res...)
		break
	}

	for _, img := range toDownload {
		if err := w.downloadImage(img); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

type image struct {
	id, slug string
}

func (w *Scraper) parseAlbum(album string) (images []image) {
	createDirIfNotExist(strings.TrimPrefix(album, "/"))

	response, err := http.Get(w.url + album)
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
			//downlink := w.url + "/download" + album + "-" + id
			images = append(images, image{
				id:   id,
				slug: slug,
			})

		}

	})

	return images
}

func (w *Scraper) downloadImage(img image) error {
	downloadURL := fmt.Sprintf("%s/download/%s-%s", w.url, img.slug, img.id)
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("user-agent", w.agent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(img.slug+"/"+img.id+".jpg", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0744)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	written, err := io.Copy(f, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Downloaded ", written, "to ", f.Name())
	return nil
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}
