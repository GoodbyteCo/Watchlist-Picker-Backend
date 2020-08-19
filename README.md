# Watchlist-Picker-Backend

Backend for [Watchlist Picker](https://watchlistpicker.com)

### Install
Requires go

#### run locally for testing
In the repo run `go run random-letterboxd.go` this will launch a localhost server at port 8080 by default but the port can be set by setting the ENV variable of PORT

### Usage

To use run `random-letterboxd USERNAME`

#### Notes

As the letterboxd API is private this scrapes the site so by the nature of webscraping is fairly slow (ex. 5 sec for 57 film watchlist) and aswell is very fragile and though I will try to keep up to date I can't make any promises that it will work. PRs welcome
