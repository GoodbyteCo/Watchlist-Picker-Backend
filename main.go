package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

//Film struct for http response
type film struct {
	Slug  string `json:"slug"`      //url of film
	Image string `json:"image_url"` //url of image
	Name  string `json:"film_name"`
}

//struct for channel to send film and whether is has finshed a user
type filmSend struct {
	film film //film to be sent over channel
	done bool //if user is done
}

const url = "https://letterboxd.com/ajax/poster" //first part of url for getting full info on film
const urlEnd = "menu/linked/125x187/"            // second part of url for getting full info on film
const site = "https://letterboxd.com"

func main() {
	getFilmHandler := http.HandlerFunc(getFilm)
	http.Handle("/film", getFilmHandler)
	log.Println("serving at :8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	http.ListenAndServe(":"+port, nil)
}

//Main handler func for request
func getFilm(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	query := r.URL.Query() //Get URL Params(type map)
	users, ok := query["users"]
	log.Println(len(users))
	if !ok || len(users) == 0 {
		http.Error(w, "no users", 400)
		return
	}
	userFilm := scrapeUser(users)
	if (userFilm == film{}) {
		http.Error(w, "no users", 404)
		return
	}
	js, err := json.Marshal(userFilm)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	w.Write(js)

}

//main scraping function
func scrapeUser(users []string) film {
	var user int = 0          //conuter for number of users increses by one when a users page starts being scraped decreses when user has finished think kinda like a semaphore
	var totalFilms []film     //final list to hold all film
	ch := make(chan filmSend) //channel to send films over
	// start go routine to scrape each user
	for _, a := range users {
		log.Println(a)
		user++
		go scrape(a, ch)
	}
	for {
		userFilm := <-ch
		if userFilm.done { //if users channel is don't then the scapre for that user has finished so decrease the user count
			user--
			if user == 0 {
				break
			}
		} else {
			totalFilms = append(totalFilms, userFilm.film) //append feilm recieved over channel to list
		}

	}

	//chose random film from list
	if len(totalFilms) == 0 {
		return film{}
	}
	rand.Seed(time.Now().UTC().UnixNano())
	n := rand.Intn(len(totalFilms))
	log.Println(len(totalFilms))
	log.Println(n)
	log.Println(totalFilms[n])
	return totalFilms[n]
}

//function to scapre an single user
func scrape(userName string, ch chan filmSend) {
	siteToVisit := site + "/" + userName + "/watchlist"

	ajc := colly.NewCollector(
		colly.Async(true),
	)
	ajc.OnHTML("div.film-poster", func(e *colly.HTMLElement) { //secondard cleector to get main data for film
		name := e.Attr("data-film-name")
		slug := e.Attr("data-target-link")
		img := e.ChildAttr("img", "src")
		tempfilm := film{
			Slug:  (site + slug),
			Image: makeBigger(img),
			Name:  name,
		}
		ch <- ok(tempfilm)
	})
	c := colly.NewCollector(
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 100})
	c.OnHTML(".poster-container", func(e *colly.HTMLElement) { //primary scarer to get url of each film that contian full information
		e.ForEach("div.film-poster", func(i int, ein *colly.HTMLElement) {
			slug := ein.Attr("data-film-slug")
			ajc.Visit(url + slug + urlEnd) //start go routine to collect all film data
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
	ajc.Wait()
	ch <- done() // users has finished so send done through channel

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

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func makeBigger(url string) string {
	return strings.ReplaceAll(url, "-0-125-0-187-", "-0-230-0-345-")
}
