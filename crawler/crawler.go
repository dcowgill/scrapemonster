package crawler

import (
	"log"
	net_url "net/url"
)

type Fetcher interface {
	Fetch(url *net_url.URL) (body string, urls []*net_url.URL, err error)
}

type URLTransformer interface {
	TransformURL(url *net_url.URL) *net_url.URL
}

type Result struct {
	URL   *net_url.URL
	Body  string
	urls  []*net_url.URL
	depth int
	err   error
}

type Crawler struct {
	Fetcher
	URLTransformer
	MaxParallel int
	MaxDepth    int
	OutputChan  chan *Result
	Verbose     bool
}

// New returns a Crawler object, using the given Fetcher implementation.
func New(f Fetcher) *Crawler {
	return &Crawler{
		Fetcher:     f,
		MaxParallel: 10,
		MaxDepth:    5,
	}
}

// Go begins crawling the website at the specified URL.
func (c *Crawler) Go(startURL string) error {
	startURL2, err := net_url.Parse(startURL)
	if err != nil {
		return err
	}

	resultChan := make(chan *Result)
	sem := make(chan int, c.MaxParallel)
	fetch := func(url *net_url.URL, depth int) {
		go func() {
			sem <- 1
			body, urls, err := c.Fetch(url)
			resultChan <- &Result{url, body, c.transformURLs(urls), depth, err}
			<-sem
		}()
	}

	go fetch(startURL2, c.MaxDepth)
	visited := map[string]bool{startURL2.String(): true}
	nprocs := 1

	for nprocs > 0 {
		r := <-resultChan
		nprocs--
		if r.depth > 0 {
			for _, url := range r.urls {
				key := url.String()
				if !visited[key] {
					go fetch(url, r.depth-1)
					visited[key] = true
					nprocs++
				}
			}
		}
		if r.err != nil {
			log.Printf("crawler: %s", r.err)
		} else if c.OutputChan != nil {
			c.OutputChan <- r
		}
	}

	if c.OutputChan != nil {
		close(c.OutputChan)
	}

	return nil
}

func (c *Crawler) transformURLs(urls []*net_url.URL) (newurls []*net_url.URL) {
	for _, u := range urls {
		if u = c.TransformURL(u); u != nil {
			newurls = append(newurls, u)
		}
	}
	return
}
