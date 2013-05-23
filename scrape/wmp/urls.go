package wmp

import (
	"fmt"
	"github.com/launchtime/scrapemonster/scrape"
	"net/url"
	"regexp"
	"strconv"
)

const HOST = "www.wemakeprice.com"

var (
	dealListPathRegexp = regexp.MustCompile(`^/(?:main|wmp_top_menu)/(\w+(?:/\d+)?)`)
	dealPathRegexp     = regexp.MustCompile(`^/deal/adeal/(\d+)`)
)

type dealListID string

func baseURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   HOST,
		Path:   "/",
	}
}

func urlForHomepage() *url.URL {
	u := baseURL()
	u.Path = "/main"
	return u
}

func urlForDealList(id dealListID) *url.URL {
	u := baseURL()
	u.Path = fmt.Sprintf("/main/%s", id)
	return u
}

func urlForGetOptionList(id scrape.DealID) *url.URL {
	u := baseURL()
	u.Path = fmt.Sprintf("/c/wmp_cart/option_layer/deal/%d", id)
	return u
}

func parseDealListURL(u *url.URL) (id dealListID, ok bool) {
	matches := matchURL(u, dealListPathRegexp)
	if matches != nil {
		id = dealListID(matches[1])
		ok = true
	}
	return
}

func parseDealURL(u *url.URL) (id scrape.DealID, ok bool) {
	matches := matchURL(u, dealPathRegexp)
	if matches != nil {
		if n, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			id = scrape.DealID(n)
			ok = true
		}
	}
	return
}

func matchURL(u *url.URL, re *regexp.Regexp) []string {
	if u != nil && u.Host == HOST {
		return re.FindStringSubmatch(u.Path)
	}
	return nil
}

func (_ *Scraper) DefaultStartURL() string {
	return urlForHomepage().String()
}

func (s *Scraper) TransformURL(u *url.URL) *url.URL {
	if id, ok := parseDealListURL(u); ok {
		return urlForDealList(id)
	}
	if id, ok := parseDealURL(u); ok {
		return s.DealURL(id)
	}
	return nil
}

func (_ *Scraper) DealURL(id scrape.DealID) *url.URL {
	u := baseURL()
	u.Path = fmt.Sprintf("/deal/adeal/%d", id)
	return u
}

func (_ *Scraper) ExtractURLs(body string) []*url.URL {
	return nil
}
