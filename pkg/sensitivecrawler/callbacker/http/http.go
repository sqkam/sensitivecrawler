package httpcallbacker

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"
)

type Option interface {
	Apply(*httpCallBacker)
}
type OptionFunc func(*httpCallBacker)

func (f OptionFunc) Apply(h *httpCallBacker) {
	f(h)
}

func WithHttpClient(c *http.Client) Option {
	return OptionFunc(func(h *httpCallBacker) {
		h.client = c
	})
}

type httpCallBacker struct {
	url    string
	client *http.Client
}

func (c *httpCallBacker) Do(ch <-chan result.Result) {
	for r := range ch {
		go c.doCallback(r)
	}
}

func (c *httpCallBacker) doCallback(r result.Result) {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(r); err != nil {
		log.Println(err.Error())
		return
	}
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, c.url, b)
	if err != nil {
		log.Println(err.Error())
		return
	}
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	_, err = c.client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func New(url string, options ...Option) callbacker.CallBacker {
	h := &httpCallBacker{
		url:    url,
		client: http.DefaultClient,
	}

	for _, o := range options {
		o.Apply(h)
	}
	return h
}
