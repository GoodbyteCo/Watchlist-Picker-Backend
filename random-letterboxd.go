package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type film struct {
	Slug  string
	Image string
	Name  string
}

const url = "https://letterboxd.com/ajax/poster"
const urlEnd = "menu/linked/125x187/"
const site = "https://letterboxd.com"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "please provide atleast one letterboxd username")
		os.Exit(1)
	}
	var totalFilms []film
	for _, a := range args {
		fmt.Println(a)
		userFilm := scrape(a)
		totalFilms = append(totalFilms, userFilm...)
	}
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(totalFilms)
	fmt.Println(len(totalFilms))
	fmt.Println(totalFilms[n])

}

func scrape(userName string) []film {
	var wg sync.WaitGroup
	siteToVisit := site + "/" + userName + "/watchlist"

	var posters []film
	ajc := colly.NewCollector()
	ajc.OnHTML("div.film-poster", func(e *colly.HTMLElement) {
		name := e.Attr("data-film-name")
		slug := e.Attr("data-target-link")
		img := e.ChildAttr("img", "src")
		tempfilm := film{
			Slug:  (site + slug),
			Image: img,
			Name:  name,
		}
		posters = append(posters, tempfilm)
		wg.Done()
	})
	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 50})
	c.OnHTML(".poster-container", func(e *colly.HTMLElement) {
		e.ForEach("div.film-poster", func(i int, ein *colly.HTMLElement) {
			slug := ein.Attr("data-film-slug")
			wg.Add(1)
			go ajc.Visit(url + slug + urlEnd)
		})
		wg.Wait()

	})
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "watchlist/page") {
			e.Request.Visit(e.Request.AbsoluteURL(link))
		}
	})

	c.Visit(siteToVisit)

	return posters
}
