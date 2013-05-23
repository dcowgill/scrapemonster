package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/launchtime/scrapemonster/cmd"
	"github.com/launchtime/scrapemonster/crawler"
	"github.com/launchtime/scrapemonster/scrape"
	"log"
	"os"
)

var (
	dealIDArg  = flag.Int("d", 0, "deal ID")
	getOptions = flag.Bool("o", true, "get deal options")
	sitename   = flag.String("s", "", "site to crawl")
)

func getDeal(s scrape.Scraper, g *crawler.Getter, id scrape.DealID) *scrape.Deal {
	url := s.DealURL(id)
	data, err := g.GetBody(url.String())
	if err != nil {
		log.Fatal(err)
	}
	deal, err := s.ParseDeal(url, string(data))
	if err != nil {
		log.Fatal(err)
	} else if deal == nil {
		log.Fatal("deal does not exist")
	}
	return deal
}

type info struct {
	Deal    *scrape.Deal
	Options []*scrape.Option
}

func main() {
	flag.Parse()

	if *dealIDArg == 0 {
		fmt.Println("Usage error: Deal ID (-d flag) is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	scraper := cmd.NewScraper(*sitename)
	dealID := scrape.DealID(*dealIDArg)
	getter := crawler.NewGetter()
	getter.Verbose = true

	info := info{Deal: getDeal(scraper, getter, dealID)}
	if *getOptions {
		info.Options = scraper.GetDealOptions(getter, dealID)
	}

	data, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(data)
	os.Stdout.WriteString("\n")
}
