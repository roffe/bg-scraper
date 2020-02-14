package desktopnexus

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
		url:   "https://www.desktopnexus.com",
		terms: terms,
	}
}

// Scrape ...
func (s *Scraper) Scrape() error {
	url := fmt.Sprintf("%s/search/%s", s.url, s.terms)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("Error loading HTTP response body: %s", err)
	}
	resp.Body.Close()

	var pages int
	document.Find("select").EachWithBreak(func(index int, element *goquery.Selection) bool {
		element.Find("option").Each(func(index int, element *goquery.Selection) {
			var err error
			pages, err = strconv.Atoi(element.Text())
			if err != nil {
				log.Println(err)
			}
		})
		return true
	})
	var images []string
	images = append(images, getImages(document)...)

	for page := 2; page <= pages; page++ {
		url2 := fmt.Sprintf("%s/%d", url, page)
		fmt.Println(url2)
		resp, err = http.Get(url2)
		if err != nil {
			return err
		}
		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("Error loading HTTP response body: %s", err)
		}
		resp.Body.Close()
		images = append(images, getImages(document)...)
	}
	semchan := make(chan struct{}, 3)
	var wg sync.WaitGroup

	if len(images) > 0 {
		utils.CreateDirIfNotExist("./download/desktopnexus/" + s.terms)
	}

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
	return nil
}

func getImages(doc *goquery.Document) []string {
	var images []string
	doc.Find("div").Each(func(index int, element *goquery.Selection) {
		if attr, exists := element.Attr("style"); exists {
			if attr == "float: left; margin: 6px;" {
				element.Children().Each(func(index int, element *goquery.Selection) {
					if href, exists := element.Attr("href"); exists {
						parts := strings.Split(strings.TrimSuffix(href, "/"), "/")
						url := fmt.Sprintf("https://videogames.desktopnexus.com/get_wallpaper_download_url.php?id=%s&w=1920&h=1200", parts[len(parts)-1])
						images = append(images, url)
					}
				})
			}
		}
	})
	return images
}

func (s *Scraper) downloadImage(img string) error {
	resp, err := http.Get(img)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	url2 := fmt.Sprintf("https:%s", body)

	u, err := url.Parse(url2)
	if err != nil {
		return err
	}
	resp, err = http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	parts := strings.Split(u.Path, "/")
	fname := parts[len(parts)-1]

	f, err := os.OpenFile("./download/desktopnexus/"+s.terms+"/"+fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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
