package wmp

type Scraper int

func (_ *Scraper) Name() string {
	return "wmp"
}
