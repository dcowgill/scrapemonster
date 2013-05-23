package coupang

import (
	"fmt"
	"github.com/launchtime/scrapemonster/scrape"
	"net/url"
	"regexp"
	"strconv"
)

var (
	extractURLRegexp = regexp.MustCompile(`(?:https?://)?(?:[-\w.%~]+)?/(?:alldeal|promotion/prmt|shopping|deal)\.pang(?:[?][-\w.%~:@!$&()*+,;=/?]*)?(?:[#][-\w.%~:@!$&()*+,;=/?]*)?`)
)

const HOST = "www.coupang.com"

func baseURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   HOST,
		Path:   "/",
	}
}

func parseDealURL(u *url.URL) (id scrape.DealID, ok bool) {
	if u.Path == "/deal.pang" {
		q := u.Query()
		if n, err := strconv.ParseInt(q.Get("coupang"), 10, 64); err == nil {
			id = scrape.DealID(n)
			ok = true
		}
	}
	return
}

func isDealListURL(u *url.URL) bool {
	switch u.Path {
	case "/alldeal.pang":
	case "/promotion/prmt.pang":
	case "/shopping.pang":
	default:
		return false
	}
	return true
}

func (_ *Scraper) DefaultStartURL() string {
	return baseURL().String()
}

func (s *Scraper) TransformURL(u *url.URL) *url.URL {
	if id, ok := parseDealURL(u); ok {
		return s.DealURL(id)
	}
	if isDealListURL(u) {
		return u
	}
	return nil
}

func (_ *Scraper) DealURL(id scrape.DealID) *url.URL {
	u := baseURL()
	u.Path = fmt.Sprintf("/deal.pang")
	q := url.Values{}
	q.Set("coupang", strconv.FormatInt(int64(id), 10))
	u.RawQuery = q.Encode()
	return u
}

func (_ *Scraper) ExtractURLs(body string) (urls []*url.URL) {
	for _, s := range extractURLRegexp.FindAllString(body, -1) {
		if u, err := url.Parse(s); err == nil {
			urls = append(urls, u)
		}
	}
	return
}
