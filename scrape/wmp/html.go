package wmp

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"fmt"
	"github.com/launchtime/scrapemonster/scrape"
	"github.com/launchtime/scrapemonster/scrape/htmlutil"
	"net/url"
	"strings"
)

var (
	categorySelector = cascadia.MustCompile("#gnb ul.gnb_menu > li.on a span.hide")

	subcategorySelector = cascadia.MustCompile("#div_section_gnbsub ul > li.on a")

	localeSelectors = []cascadia.Selector{
		cascadia.MustCompile(".gnb_section.region .gnb_sub h3.on a"),
		cascadia.MustCompile(".gnb_section.region .gnb_sub ul > li.on a"),
	}

	numSoldSelector = cascadia.MustCompile("#buy_num")

	originalPriceSelector = cascadia.MustCompile(
		".price_area .ba_origin_price")

	discountPriceSelector = cascadia.MustCompile(
		".price_area .ba_sale_price")
)

type dealPage struct {
	dealID scrape.DealID
	body   string
	root   *html.Node
}

func newDealPage(id scrape.DealID, body string) (p *dealPage, err error) {
	var root *html.Node
	root, err = html.Parse(strings.NewReader(body))
	if err != nil {
		return
	}
	p = &dealPage{id, body, root}
	return
}

func (p *dealPage) description() *string {
	selector := cascadia.MustCompile(fmt.Sprintf("img#img_onecut_%d", p.dealID))
	nodes := selector.MatchAll(p.root)
	if len(nodes) == 1 {
		if attr := htmlutil.GetAttr(nodes[0], "alt"); attr != nil {
			s := attr.Val
			return &s
		}
	}
	return nil
}

func (p *dealPage) category() *string {
	nodes := categorySelector.MatchAll(p.root)
	if len(nodes) == 1 {
		return htmlutil.FirstText(nodes[0])
	}
	return nil
}

func (p *dealPage) subcategory() *string {
	nodes := subcategorySelector.MatchAll(p.root)
	if len(nodes) == 1 {
		return htmlutil.FirstText(nodes[0])
	}
	return nil
}

func (p *dealPage) locale() []string {
	var locale []string
	for _, sel := range localeSelectors {
		nodes := sel.MatchAll(p.root)
		if len(nodes) == 1 {
			if s := htmlutil.FirstText(nodes[0]); s != nil {
				locale = append(locale, *s)
			}
		}
	}
	return locale
}

func (p *dealPage) originalPrice() *int {
	return htmlutil.ExtractInteger(p.root, originalPriceSelector)
}

func (p *dealPage) discountPrice() *int {
	return htmlutil.ExtractInteger(p.root, discountPriceSelector)
}

func (p *dealPage) numSold() *int {
	return htmlutil.ExtractInteger(p.root, numSoldSelector)
}

func (p *dealPage) expired() bool {
	return false
}

func (p *dealPage) adult() bool {
	return false
}

func (s *Scraper) ParseDeal(u *url.URL, body string) (d *scrape.Deal, err error) {
	dealID, ok := parseDealURL(u)
	if !ok {
		return
	}
	p, err := newDealPage(dealID, body)
	if err != nil {
		return
	}
	d = &scrape.Deal{
		SiteName:      s.Name(),
		DealID:        dealID,
		Description:   p.description(),
		Category:      p.category(),
		Subcategory:   p.subcategory(),
		Locale:        p.locale(),
		OriginalPrice: p.originalPrice(),
		DiscountPrice: p.discountPrice(),
		NumSold:       p.numSold(),
		Expired:       p.expired(),
		Adult:         p.adult(),
	}
	return
}
