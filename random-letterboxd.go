package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type film struct {
	Slug  string `json:"slug"`
	Image string `json:"image_url"`
	Name  string `json:"film_name"`
}

type filmSend struct {
	film film
	ok   bool
}

const url = "https://letterboxd.com/ajax/poster"
const urlEnd = "menu/linked/125x187/"
const site = "https://letterboxd.com"

func main() {
	getFilmHandler := http.HandlerFunc(getFilm)
	http.Handle("/film", getFilmHandler)
	fmt.Println("serving at :8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	http.ListenAndServe(":"+port, nil)
}

func getFilm(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	query := r.URL.Query()
	users, ok := query["users"]
	if !ok || len(users) == 0 {
		http.Error(w, "no users", 400)
	}
	fmt.Println(users)
	userFilm := scrapeUser(users)
	if (userFilm == film{}) {
		http.Error(w, "no users", 404)
	}
	js, err := json.Marshal(userFilm)
	if err != nil {
		http.Error(w, "internal error", 500)
	}
	w.Write(js)

}

func scrapeUser(users []string) film {
	var user int = 0
	var totalFilms []film
	ch := make(chan filmSend)
	for _, a := range users {
		fmt.Println(a)
		user++
		go scrape(a, ch)
	}
	for {
		userFilm := <-ch
		if userFilm.ok == false {
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
		return film{}
	}
	n := rand.Intn(len(totalFilms))
	log.Println(len(totalFilms))
	log.Println(n)
	log.Println(totalFilms[n])
	return totalFilms[n]
}

func scrape(userName string, ch chan filmSend) {
	var wg sync.WaitGroup
	siteToVisit := site + "/" + userName + "/watchlist"

	ajc := colly.NewCollector()
	ajc.OnHTML("div.film-poster", func(e *colly.HTMLElement) {
		name := e.Attr("data-film-name")
		slug := e.Attr("data-target-link")
		img := e.ChildAttr("img", "src")
		tempfilm := film{
			Slug:  (site + slug),
			Image: makeBigger(img),
			Name:  name,
		}
		ch <- ok(tempfilm)
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
	ch <- done()

}

func ok(f film) filmSend {
	return filmSend{
		film: f,
		ok:   true,
	}
}

func done() filmSend {
	return filmSend{
		film: film{},
		ok:   false,
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func makeBigger(url string) string {
	return strings.ReplaceAll(url, "-0-125-0-187-", "-0-230-0-345-")
}
