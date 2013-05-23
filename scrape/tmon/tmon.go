package tmon

type Scraper int

func (_ *Scraper) Name() string {
	return "tmon"
}
