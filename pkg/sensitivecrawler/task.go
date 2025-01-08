package sensitivecrawler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gocolly/colly/v2"
	"github.com/imthaghost/goclone/pkg/parser"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivematcher"
)

type TaskOption interface {
	Apply(*task)
}
type TaskOptionOptionFunc func(*task)

func (f TaskOptionOptionFunc) Apply(t *task) {
	f(t)
}

func WithCallBacker(c callbacker.CallBacker) TaskOption {
	return TaskOptionOptionFunc(func(t *task) {
		t.callBacker = c
	})
}

func WithMaxDepth(d int) TaskOption {
	return TaskOptionOptionFunc(func(t *task) {
		t.c.MaxDepth = d
	})
}

// MaxBodySize  UserAgent AllowedDomains URLFilters MaxBodySize CacheDir
func WithCollyCollector(c *colly.Collector) TaskOption {
	return TaskOptionOptionFunc(func(t *task) {
		c.Async = true
		t.c = c
	})
}

// 并发数
func WithLimitRules(rules []*colly.LimitRule) TaskOption {
	return TaskOptionOptionFunc(func(h *task) {
		h.c.Limits(rules)
	})
}

func WithTimeOut(t int64) TaskOption {
	return TaskOptionOptionFunc(func(h *task) {
		h.timeout = t
	})
}

type task struct {
	site           string
	callBacker     callbacker.CallBacker
	m              sensitivematcher.SensitiveMatcher
	resultMsgCh    chan result.Result
	c              *colly.Collector
	urlCount       int64
	sensitiveCount int64
	timeout        int64
}

func (t *task) Analyze(ctx context.Context, url string) {
	fmt.Println("Analyzing --> ", url)
	// get the html body

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	// Closure
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err == nil {
		matchStrings := t.m.Match(respBody, url)
		if len(matchStrings) > 0 {
			atomic.AddInt64(&t.sensitiveCount, int64(len(matchStrings)))
			for _, matchStr := range matchStrings {
				t.resultMsgCh <- result.Result{Url: url, Info: matchStr}
			}

		}

	}
}

func (t *task) HtmlAnalyze(ctx context.Context, url string) {
	fmt.Println("Analyzing --> ", url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// get the html body
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	// Closure
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err == nil {
		matchStrings := t.m.Match(respBody, url)
		if len(matchStrings) > 0 {
			atomic.AddInt64(&t.sensitiveCount, int64(len(matchStrings)))
			for _, matchStr := range matchStrings {
				t.resultMsgCh <- result.Result{Url: url, Info: matchStr}
			}

		}
	}
}

func (t *task) Run(ctx context.Context) {
	url := t.site
	// url := "https://zgo.sqkam.cfd"
	isValid, isValidDomain := parser.ValidateURL(url), parser.ValidateDomain(url)
	if !isValid && !isValidDomain {
		fmt.Printf("%q is not valid", url)
		return
	}

	name := url
	if isValidDomain {
		url = parser.CreateURL(name)
	} else {
		name = parser.GetDomain(url)
	}

	// search for all link tags that have a rel attribute that is equal to stylesheet - CSS
	t.c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		// hyperlink reference
		link := e.Attr("href")
		// print css file was found
		fmt.Println("Css found", "-->", link)
		// extraction
		t.Analyze(ctx, e.Request.AbsoluteURL(link))
	})

	// search for all script tags with src attribute -- JS
	t.c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		// src attribute
		link := e.Attr("src")
		// Print link
		// fmt.Println("Js found", "-->", link)
		// extraction
		t.Analyze(ctx, e.Request.AbsoluteURL(link))
	})

	// serach for all img tags with src attribute -- Images
	t.c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		// src attribute
		link := e.Attr("src")
		if strings.HasPrefix(link, "data:image") || strings.HasPrefix(link, "blob:") {
			return
		}
		// Print link
		fmt.Println("Img found", "-->", link)
		// extraction
		t.Analyze(ctx, e.Request.AbsoluteURL(link))
	})

	// Before making a request
	t.c.OnRequest(func(r *colly.Request) {
		atomic.AddInt64(&t.urlCount, 1)
		link := r.URL.String()
		if url == link {
			t.HtmlAnalyze(ctx, link)
		}
	})

	// Visit each url and wait for stuff to load :)
	if err := t.c.Visit(url); err != nil {
		return
	}
	t.c.Wait()
}
