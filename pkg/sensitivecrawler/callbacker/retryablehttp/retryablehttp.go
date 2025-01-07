package http

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

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

func WithRetryMax(c int64) Option {
	return OptionFunc(func(h *httpCallBacker) {
		h.retryMax = c
	})
}

func WithRetryInterval(t time.Duration) Option {
	return OptionFunc(func(h *httpCallBacker) {
		h.retryInterval = t
	})
}

type httpCallBacker struct {
	url           string
	client        *http.Client
	retryMax      int64
	retryInterval time.Duration
}

func (c *httpCallBacker) Do(ch <-chan result.Result) {
	for r := range ch {
		go c.doCallback(r)
	}
}

func (c *httpCallBacker) doCallbackOnce(r result.Result) error {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(r); err != nil {
		return err
	}
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, c.url, b)
	if err != nil {
		return err
	}
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	_, err = c.client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *httpCallBacker) doCallback(r result.Result) {
	for range c.retryMax {
		err := c.doCallbackOnce(r)
		if err == nil {
			return
		}
		log.Println(err.Error())
		time.Sleep(c.retryInterval)
	}
}

func NewHttpCallBacker(url string, options ...Option) callbacker.CallBacker {
	h := &httpCallBacker{
		url:           url,
		client:        http.DefaultClient,
		retryMax:      5,
		retryInterval: 3 * time.Second,
	}

	for _, o := range options {
		o.Apply(h)
	}
	return h
}
