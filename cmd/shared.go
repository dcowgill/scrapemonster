package cmd

import (
	"github.com/launchtime/scrapemonster/scrape"
	"github.com/launchtime/scrapemonster/scrape/coupang"
	"github.com/launchtime/scrapemonster/scrape/tmon"
	"github.com/launchtime/scrapemonster/scrape/wmp"
	"log"
	"os"
)

func GetMySQLConnectionURI() string {
	if uri := os.Getenv("MYSQL_CONNECTION_URI"); uri != "" {
		return uri
	}
	return "coupang//"
}

func NewScraper(site string) scrape.Scraper {
	switch site {
	case "coupang":
		return new(coupang.Scraper)
	case "tmon":
		return new(tmon.Scraper)
	case "wmp":
		return new(wmp.Scraper)
	}
	log.Fatalf(`could not create scraper: invalid site "%s"`, site)
	return nil
}
