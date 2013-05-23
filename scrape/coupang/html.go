package coupang

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"github.com/launchtime/scrapemonster/scrape"
	"github.com/launchtime/scrapemonster/scrape/htmlutil"
	"net/url"
	"regexp"
	"strings"
)

var (
	gnbCategorySelector = cascadia.MustCompile(`#gnbDepth1 > .on`)

	gnbSubcategorySelector = cascadia.MustCompile(`#gnbTopSubMenu > .on`)

	gnbLocaleSelector = cascadia.MustCompile(`#localCatePos .on`)

	titleSelector = cascadia.MustCompile(`title`)

	numSoldSelector = cascadia.MustCompile(`#buyCount`)

	originalPriceSelector = cascadia.MustCompile(`.priceArea .originPrice .delPrice`)

	discountPriceSelector = cascadia.MustCompile(`.priceArea .salePrice`)

	expiredBuyButtonSelector = cascadia.MustCompile(`#non_click_order_button`)

	adultWarningSelector = cascadia.MustCompile(`#onlyAdult`)

	adultWarningRegexp = regexp.MustCompile(`청소년보호법`)
)

var catmap = map[string]string{
	"menuTab1": "오늘의 추천", // today
	"menuTab2": "지역",     // local
	"menuTab3": "쇼핑",     // shopping
	"menuTab4": "여행/레저",  // travel/leisure
	"menuTab5": "문화",     // culture
	"menuTab6": "오늘마감",   // closing today
	"menuTab7": "전체보기",   // all
}

var subcatmap = map[string]string{
	"gts51": "전국/서울",     // Nationwide / Seoul
	"gts52": "인천/경기",     // Incheon / Gyeong-gi
	"gts53": "대구/부산",     // Daegu / Busan
	"gts54": "대전/광주",     // Daejeon / Guangzhou
	"gts55": "강원/제주",     // Kangwon / Jeju
	"gts1":  "쇼핑 스페셜",    // special
	"gts2":  "의류",        // clothing
	"gts3":  "패션잡화",      // fashion accessories
	"gts4":  "스포츠/레저",    // sports/leisure
	"gts5":  "신품",        // new
	"gts6":  "뷰티",        // beauty
	"gts7":  "생활/주방",     // living/kitchen
	"gts8":  "홈 인테리어/취미", // home interior/hobby
	"gts9":  "디지탈/가전",    // digital/appliances
	"gts10": "출산/유아동",    // birth/baby & child
	"gts11": "쇼핑몰 할인권",   // shopping mall vouchers
	"gts31": "전체",        // full
	"gts32": "해외",        // overseas
	"gts33": "국내",        // domestic
	"gts34": "제주",        // jeju
	"gts35": "레저/입장권",    // leisure / tickets
	"gts36": "숙박",        // accommodation
	"gts41": "전지역",       // all
	"gts42": "서울/경기",     // seoul/gyeonggi
	"gts43": "다른 지역",     // other regions
}

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

func (p *dealPage) exists() bool {
	if s := p.description(); s != nil {
		return !strings.Contains(*s, "유효하지")
	}
	return false
}

func (p *dealPage) description() *string {
	nodes := titleSelector.MatchAll(p.root)
	if len(nodes) == 1 {
		return htmlutil.FirstText(nodes[0])
	}
	return nil
}

func (p *dealPage) category() *string {
	nodes := gnbCategorySelector.MatchAll(p.root)
	if len(nodes) == 1 {
		if attr := htmlutil.GetAttr(nodes[0], "id"); attr != nil {
			if s, ok := catmap[attr.Val]; ok {
				return &s
			}
			s := attr.Val
			return &s
		}
	}
	return nil
}

func (p *dealPage) subcategory() *string {
	nodes := gnbSubcategorySelector.MatchAll(p.root)
	if len(nodes) == 1 {
		if attr := htmlutil.GetAttr(nodes[0], "id"); attr != nil {
			if s, ok := subcatmap[attr.Val]; ok {
				return &s
			}
			s := attr.Val
			return &s
		}
	}
	return nil
}

func (p *dealPage) locale() []string {
	nodes := gnbLocaleSelector.MatchAll(p.root)
	if len(nodes) > 0 {
		locale := make([]string, 0, len(nodes))
		for _, n := range nodes {
			if s := htmlutil.FirstText(n); s != nil && *s != "" {
				locale = append(locale, *s)
			}
		}
		return locale
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
	return htmlutil.ExtractInteger(p.root, numSoldSelector)
}

func (p *dealPage) expired() bool {
	nodes := expiredBuyButtonSelector.MatchAll(p.root)
	return len(nodes) != 0
}

func (p *dealPage) adult() bool {
	return len(adultWarningSelector.MatchAll(p.root)) >= 1 &&
		adultWarningRegexp.FindStringSubmatch(p.body) != nil
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
