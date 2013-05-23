package tmon

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/launchtime/scrapemonster/crawler"
	"github.com/launchtime/scrapemonster/scrape"
	"log"
	"strings"
)

type Getter interface {
	GetBody(url string) (body []byte, err error)
}

type fuzzyString string

func (fs *fuzzyString) UnmarshalJSON(b []byte) (err error) {
	var (
		s string
		i int
	)
	if err = json.Unmarshal(b, &s); err == nil {
		*fs = fuzzyString(s)
	} else if err = json.Unmarshal(b, &i); err == nil {
		*fs = fuzzyString(fmt.Sprintf("%d", i))
	}
	return
}

func (fs *fuzzyString) String() string {
	if fs != nil {
		return string(*fs)
	}
	return ""
}

type rawOption struct {
	// AlwaysSale       *string       `json:"always_sale"`
	// BuywaitAvail     int           `json:"buywait_avail"`
	// Count            int           `json:"count"`
	// DealAmount       int           `json:"deal_amount"`
	DealBuyCount int   `json:"deal_buy_count"`
	DealSRL      int64 `json:"deal_srl"`
	// DealType         string        `json:"deal_type"`
	// DeliveryAmount   int           `json:"delivery_amount"`
	// DeliveryIfAmount int           `json:"delivery_if_amount"`
	// DeliveryPolicy   *string       `json:"delivery_policy"`
	FuzzyKey *fuzzyString `json:"key"`
	// MaxBuyCount      int           `json:"max_buy_count"`
	// MaxPrice         int           `json:"max_price"`
	// MinPrice         int           `json:"min_price"`
	// OneManMaxCount   int           `json:"one_man_max_count"`
	Opts string `json:"opts"`
	// OverDate         *string       `json:"over_date"`
	Price       int `json:"price"`
	RemainCount int `json:"remain_count"`
	// ReserveMaxCount  int           `json:"reserve_max_count"`
	// StatusType       string        `json:"status_type"`
	dealID   scrape.DealID `json:"-"`
	depth    int           `json:"-"`
	maxDepth int           `json:"-"`
	parent   *rawOption    `json:"-"`
}

func (o *rawOption) key() string {
	return o.FuzzyKey.String()
}

func (o *rawOption) optKey() string {
	key := o.key() + "|"
	for p := o.parent; p != nil; p = p.parent {
		key = p.key() + "|" + key
	}
	return key
}

func unmarshalOptions(body []byte) (options []*rawOption, err error) {
	// Try to parse the options JSON into map data structure.
	var optMap map[string]*rawOption
	err = json.Unmarshal(body, &optMap)
	if err == nil {
		// Success: copy the map's options into an array.
		for _, o := range optMap {
			options = append(options, o)
		}
		return
	}
	// If we encountered a type error, retry with an array.
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		err = json.Unmarshal(body, &options)
	}
	return
}

func getOptions(g *crawler.Getter, dealID scrape.DealID, parent *rawOption) (options []*rawOption) {
	var (
		body   []byte
		err    error
		depth  int
		optKey string
	)

	if parent != nil {
		depth = parent.depth + 1
		optKey = parent.optKey()
	}

	url := urlForGetOptionList(dealID, depth, optKey).String()
	body, err = g.GetBody(url)
	if err != nil {
		log.Print(err.Error())
		return
	}

	options, err = unmarshalOptions(body)
	if err != nil {
		log.Print(err.Error())
	}

	for _, o := range options {
		o.dealID = dealID
		o.depth = depth
		o.maxDepth = strings.Count(o.Opts, "|")
		o.parent = parent
	}
	return
}

func (s *Scraper) GetDealOptions(g *crawler.Getter, id scrape.DealID) []*scrape.Option {
	var (
		options = make([]*scrape.Option, 0)
		q       = list.New()
	)
	enqueueOptions := func(opts []*rawOption) {
		for _, o := range opts {
			q.PushBack(o)
		}
	}
	enqueueOptions(getOptions(g, id, nil))
	for q.Front() != nil {
		rawopt := q.Remove(q.Front()).(*rawOption)
		if rawopt.depth < rawopt.maxDepth {
			enqueueOptions(getOptions(g, id, rawopt))
		} else {
			o := &scrape.Option{
				SiteName:     s.Name(),
				DealID:       rawopt.dealID,
				OptionID:     scrape.OptionID(rawopt.DealSRL),
				Description:  rawopt.optKey(),
				Price:        rawopt.Price,
				NumAvailable: rawopt.RemainCount,
				NumSold:      rawopt.DealBuyCount,
			}
			options = append(options, o)
		}
	}
	return options
}
