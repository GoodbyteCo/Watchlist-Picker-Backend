package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/pkg/browser"
)

type film struct {
	Slug  string
	Image string
	Name  string
}

type filmSend struct {
	film film
	done bool
}

const url = "https://letterboxd.com/ajax/poster"
const urlEnd = "menu/linked/125x187/"
const site = "https://letterboxd.com"

func main() {
	args := os.Args[1:]
	var user int = 0
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "please provide atleast one letterboxd username")
		os.Exit(1)
	}
	var totalFilms []film
	ch := make(chan filmSend)
	for _, a := range args {
		fmt.Println(a)
		user++
		go scrape(a, ch)
	}
	for {
		userFilm := <-ch
		if userFilm.done {
			user--
			if user == 0 {
				break
			}
		} else {
			totalFilms = append(totalFilms, userFilm.film)
		}

	}
	rand.Seed(time.Now().Unix())
	if len(totalFilms) == 0 {
		return
	}
	n := rand.Int() % len(totalFilms)
	fmt.Println(len(totalFilms))
	fmt.Println(totalFilms[n])
	browser.OpenURL(totalFilms[n].Slug)
}

func scrape(userName string, ch chan filmSend) {
	siteToVisit := site + "/" + userName + "/watchlist"

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
		ch <- ok(tempfilm)
	})
	c := colly.NewCollector(
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 100})
	c.OnHTML(".poster-container", func(e *colly.HTMLElement) {
		e.ForEach("div.film-poster", func(i int, ein *colly.HTMLElement) {
			slug := ein.Attr("data-film-slug")
			ajc.Visit(url + slug + urlEnd)
		})

	})
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "watchlist/page") {
			e.Request.Visit(e.Request.AbsoluteURL(link))
		}
	})

	c.Visit(siteToVisit)
	c.Wait()
	ch <- done()

}

func ok(f film) filmSend {
	return filmSend{
		film: f,
		done: false,
	}
}

func done() filmSend {
	return filmSend{
		film: film{},
		done: true,
	}
}
