package scrape

import (
	"github.com/launchtime/scrapemonster/crawler"
	"net/url"
	"strconv"
)

type (
	DealID   int64
	OptionID int64

	Deal struct {
		SiteName      string
		DealID        DealID
		Description   *string
		Category      *string
		Subcategory   *string
		Locale        []string
		OriginalPrice *int
		DiscountPrice *int
		NumSold       *int
		Expired       bool
		Adult         bool
	}

	Option struct {
		SiteName     string
		DealID       DealID
		OptionID     OptionID
		Description  string
		Price        int
		NumAvailable int
		NumSold      int
	}
)

func (id DealID) MarshalJSON() (data []byte, err error) {
	return []byte(strconv.FormatInt(int64(id), 10)), nil
}

func (id OptionID) MarshalJSON() (data []byte, err error) {
	return []byte(strconv.FormatInt(int64(id), 10)), nil
}

type Scraper interface {
	Name() string
	DefaultStartURL() string
	TransformURL(u *url.URL) *url.URL
	DealURL(id DealID) *url.URL
	ParseDeal(u *url.URL, body string) (*Deal, error)
	GetDealOptions(g *crawler.Getter, id DealID) []*Option

	// This method satisfies the crawler.URLExtractor interface.
	ExtractURLs(body string) []*url.URL
}
