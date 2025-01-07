package callbacker

import (
	"net/http"

	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"
)

type Option interface {
	Apply(*retryAbleHttpCallBacker)
}
type OptionFunc func(*retryAbleHttpCallBacker)

func (f OptionFunc) Apply(h *retryAbleHttpCallBacker) {
	f(h)
}

func WithHttpClient(c *http.Client) Option {
	return OptionFunc(func(h *retryAbleHttpCallBacker) {
		h.client = c
	})
}

type retryAbleHttpCallBacker struct {
	url    string
	client *http.Client
}

func (c *retryAbleHttpCallBacker) Do(ch <-chan result.Result) {
	for r := range ch {
		go c.doCallback(r)
	}
}

func (c *retryAbleHttpCallBacker) doCallback(r result.Result) {
}

func NewRetryAbleHttpCallBacker(url string, options ...Option) CallBacker {
	h := &retryAbleHttpCallBacker{
		url:    url,
		client: http.DefaultClient,
	}

	for _, o := range options {
		o.Apply(h)
	}
	return h
}
