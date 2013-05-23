package crawler

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	net_url "net/url"
)

// URLExtractor enables customized URL extraction in a fetcher.
type URLExtractor interface {
	ExtractURLs(body string) []*net_url.URL
}

// SimpleFetcher is a default implementation of the Fetcher interface that is
// suitable for most crawlers.
type SimpleFetcher struct {
	*Getter
	URLExtractor
}

// Fetch requests the specified URL and returns the response body plus all
// links URLs found in the body (see also parseLinks). Only returns URLs whose
// hostname exactly matches the hostname of the source URL.
func (f SimpleFetcher) Fetch(url *net_url.URL) (body string, urls []*net_url.URL, err error) {
	// Fetch the url body.
	var data []byte
	data, err = f.GetBody(url.String())
	if err != nil {
		return
	}
	body = string(data)
	// Extract urls from <a> elements in html body.
	urls = parseLinks(body)
	// Add urls found by our custom link extractor.
	for _, u := range f.ExtractURLs(body) {
		urls = append(urls, u)
	}
	// Resolve relative urls and use a map to remove duplicates.
	urlmap := make(map[string]*net_url.URL, 0)
	for _, u := range urls {
		u = url.ResolveReference(u)
		if u.Host == url.Host {
			urlmap[u.String()] = u
		}
	}
	// Reuse the storage pointed to by urls.
	urls = urls[:0]
	for _, u := range urlmap {
		urls = append(urls, u)
	}
	return
}

// Find all tags matching <a href="..."> in the given HTML body and returns
// their href attributes as URL objects.
func parseLinks(body string) (urls []*net_url.URL) {
	r := bytes.NewBufferString(body)
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			return
		}
		if tt == html.StartTagToken {
			tok := z.Token()
			if tok.DataAtom == atom.A {
				for _, attr := range tok.Attr {
					if attr.Key == "href" {
						if u, err := net_url.Parse(attr.Val); err == nil {
							urls = append(urls, u)
						}
					}
				}
			}
		}
	}
	return
}
