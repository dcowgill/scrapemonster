package tmon

import (
	"fmt"
	"github.com/launchtime/scrapemonster/scrape"
	"net/url"
	"regexp"
	"strconv"
)

const HOST = "www.ticketmonster.co.kr"

var (
	dealListPathRegexp = regexp.MustCompile(`^/deallist/(\d+)`)
	dealPathRegexp     = regexp.MustCompile(`^/deal/(\d+)`)
)

type dealListID int64

func baseURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   HOST,
		Path:   "/",
	}
}

func urlForDealList(id dealListID) *url.URL {
	u := baseURL()
	u.Path = fmt.Sprintf("/deallist/%d", id)
	return u
}

func urlForGetOptionList(id scrape.DealID, depth int, optKey string) *url.URL {
	u := baseURL()
	u.Path = fmt.Sprintf("/deal/getOptionList/%d/%d", id, depth)
	q := u.Query()
	q.Set("opt_key", optKey)
	u.RawQuery = q.Encode()
	return u
}

func parseDealListURL(u *url.URL) (id dealListID, ok bool) {
	matches := matchURL(u, dealListPathRegexp)
	if matches != nil {
		if n, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			id = dealListID(n)
			ok = true
		}
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
	u := baseURL()
	u.Path = "/home/"
	return u.String()
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
	u.Path = fmt.Sprintf("/deal/%d", id)
	return u
}

func (_ *Scraper) ExtractURLs(body string) []*url.URL {
	return nil
}
