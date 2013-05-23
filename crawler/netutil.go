package crawler

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

var (
	UserAgentStrings = map[string]string{
		"MSIE8": `Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0)`,
	}
)

type Getter struct {
	UserAgent string
	Timeout   time.Duration
	Verbose   bool
	transport *http.Transport
}

func NewGetter() *Getter {
	g := new(Getter)
	g.transport = &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			if g.Timeout.Nanoseconds() > 0 {
				return net.DialTimeout(network, addr, g.Timeout)
			}
			return net.Dial(network, addr)
		},
	}
	return g
}

func (g *Getter) GetBody(url string) (data []byte, err error) {
	// Build a map of HTTP headers.
	headers := make(map[string]string)
	if g.UserAgent != "" {
		headers["User-Agent"] = g.UserAgent
	}

	// Create the request and add our headers.
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Send the request and read the response body.
	client := &http.Client{Transport: g.transport}
	if g.Verbose {
		log.Printf("GET %s", url)
	}
	var rsp *http.Response
	rsp, err = client.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()
	data, err = ioutil.ReadAll(rsp.Body)
	return
}
