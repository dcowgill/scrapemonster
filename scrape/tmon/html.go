package tmon

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"github.com/launchtime/scrapemonster/scrape"
	"github.com/launchtime/scrapemonster/scrape/htmlutil"
	net_url "net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	titleSelector = cascadia.MustCompile(`head title`)

	gnbActiveSectionTabsSelector = cascadia.MustCompile(
		`div.gnb_section ul.tab_gnb > li.on > a`)

	gnbActiveSubmenuSelector = cascadia.MustCompile(
		`div.gnb_section div.submenu ul > li.on > a`)

	gnbLocaleRegexp = regexp.MustCompile(
		`return \(/\(\\b(\d+)\\b\)/.test\(this.href\)\);`)

	gnbLocaleSelector = cascadia.MustCompile(`div.gnb_section.local ul > li > a`)

	numSoldRegexp = regexp.MustCompile(`countUpTo\((\d+)\)`)

	originalPriceSelector = cascadia.MustCompile(`.price_info .price .old em`)

	discountPriceSelector = cascadia.MustCompile(`.price_info .price .now_price`)

	buyButtonSelector = cascadia.MustCompile(`a#buy_button`)

	adultWarningSelector = cascadia.MustCompile(`div#content div.deal_detail_adult`)

	adultWarningRegexp = regexp.MustCompile(`청소년보호법`)

	notFoundErrorSelector = cascadia.MustCompile(`.error_type .no_find`)
)

type dealPage struct {
	body string
	root *html.Node
}

func newDealPage(body string) (p *dealPage, err error) {
	var root *html.Node
	root, err = html.Parse(strings.NewReader(body))
	if err != nil {
		return
	}
	p = &dealPage{body: body, root: root}
	return
}

func (p *dealPage) exists() bool {
	return len(notFoundErrorSelector.MatchAll(p.root)) == 0
}

func (p *dealPage) description() *string {
	nodes := titleSelector.MatchAll(p.root)
	if len(nodes) == 1 {
		return htmlutil.FirstText(nodes[0])
	}
	return nil
}

func (p *dealPage) category() *string {
	nodes := gnbActiveSectionTabsSelector.MatchAll(p.root)
	if len(nodes) == 1 {
		return htmlutil.FirstText(nodes[0])
	}
	return nil
}

func (p *dealPage) subcategory() *string {
	nodes := gnbActiveSubmenuSelector.MatchAll(p.root)
	if len(nodes) == 1 {
		return htmlutil.FirstText(nodes[0])
	}
	return nil
}

func (p *dealPage) locale() []string {
	// Look for the javascript code that highlights the local subsection.
	matches := gnbLocaleRegexp.FindAllStringSubmatch(p.body, -1)
	if len(matches) != 1 {
		return nil
	}
	dealListID, err := strconv.ParseInt(matches[0][1], 10, 64)
	if err != nil {
		return nil
	}

	// For every link in the local subsection of the global nav:
	baseURL := baseURL()
	nodes := gnbLocaleSelector.MatchAll(p.root)
	for _, node := range nodes {
		// Find the href attribute.
		if attr := htmlutil.GetAttr(node, "href"); attr != nil {
			// Parse the href and make it an absolute URL.
			if url, err := net_url.Parse(attr.Val); err == nil {
				url = baseURL.ResolveReference(url)
				// If the link points to our deal list page...
				if id, ok := parseDealListURL(url); ok && int64(id) == dealListID {
					// ...return the link's text node.
					if s := htmlutil.FirstText(node); s != nil {
						return []string{*s}
					}
					return nil
				}
			}
		}
	}
	return nil
}

func (p *dealPage) originalPrice() *int {
	return htmlutil.ExtractInteger(p.root, originalPriceSelector)
}

func (p *dealPage) discountPrice() *int {
	return htmlutil.ExtractInteger(p.root, discountPriceSelector)
}

func (p *dealPage) numSold() *int {
	matches := numSoldRegexp.FindAllStringSubmatch(p.body, -1)
	if len(matches) >= 1 {
		s := matches[len(matches)-1][1]
		if i, err := strconv.Atoi(s); err == nil {
			return &i
		}
	}
	return nil
}

func (p *dealPage) expired() bool {
	nodes := buyButtonSelector.MatchAll(p.root)
	if len(nodes) == 1 {
		s := htmlutil.FirstText(nodes[0])
		return s != nil && *s == "판매종료"
	}
	return false
}

func (p *dealPage) adult() bool {
	return len(adultWarningSelector.MatchAll(p.root)) >= 1 &&
		adultWarningRegexp.FindStringSubmatch(p.body) != nil
}

func (s *Scraper) ParseDeal(u *net_url.URL, body string) (d *scrape.Deal, err error) {
	dealID, ok := parseDealURL(u)
	if !ok {
		return
	}
	p, err := newDealPage(body)
	if err != nil {
		return
	}
	if !p.exists() {
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
