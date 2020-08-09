# random-letterboxd

A Command line tool to pick a movie for you to watch from your letterboxd watch list

### Install

#### Without Go
To install without go clone and then run the build.sh script

#### With Go
run `go get github.com/holopollock/random-letterboxd`

### Usage

To use run `random-letterboxd USERNAME`

#### Notes

As the letterboxd API is private this scrapes the site so by the nature of webscraping is fairly slow (ex. 5 sec for 57 film watchlist) and aswell is very fragile and though I will try to keep up to date I can't make any promises that it will work. PRs welcome
