package coupang

import (
	"github.com/launchtime/scrapemonster/crawler"
	"github.com/launchtime/scrapemonster/scrape"
)

type Scraper int

func (_ *Scraper) Name() string {
	return "coupang"
}

func (s *Scraper) GetDealOptions(g *crawler.Getter, id scrape.DealID) []*scrape.Option {
	return nil
}
