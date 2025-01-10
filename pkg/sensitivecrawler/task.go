package sensitivecrawler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/fatih/color"

	"github.com/gocolly/colly/v2"
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

func WithAllowedDomains(d []string) TaskOption {
	return TaskOptionOptionFunc(func(t *task) {
		t.c.AllowedDomains = append(t.c.AllowedDomains, d...)
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
	timeout        int64
	urlCount       int64
	analyzeCount   int64
	analyzeBytes   int64
	sensitiveCount int64
}

func (t *task) Analyze(ctx context.Context, url string) {
	color.New(color.FgYellow).Println("Analyzing --> ", url)
	// get the html body

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Println("Analyze NewRequestWithContext error ", err.Error())
		return

	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Analyze  http.DefaultClient.Do error ", err.Error())
		return

	}
	// Closure
	defer resp.Body.Close()
	atomic.AddInt64(&t.analyzeCount, 1)

	respBody, err := io.ReadAll(resp.Body)
	if err == nil {
		atomic.AddInt64(&t.analyzeBytes, int64(len(respBody)))
		matchStrings := t.m.Match(ctx, respBody)
		if len(matchStrings) > 0 {
			atomic.AddInt64(&t.sensitiveCount, int64(len(matchStrings)))
			for _, matchStr := range matchStrings {
				t.resultMsgCh <- result.Result{Url: url, Info: matchStr}
			}

		}
	}
}

func (t *task) HtmlAnalyze(ctx context.Context, url string) {
	color.New(color.FgYellow).Println("Analyzing --> ", url)
	//
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// get the html body
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Println("HtmlAnalyze NewRequestWithContext error ", err.Error())
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("HtmlAnalyze  http.DefaultClient.Do error ", err.Error())
		return
	}
	// Closure
	defer resp.Body.Close()
	atomic.AddInt64(&t.analyzeCount, 1)

	respBody, err := io.ReadAll(resp.Body)
	if err == nil {
		atomic.AddInt64(&t.analyzeBytes, int64(len(respBody)))
		matchStrings := t.m.Match(ctx, respBody)
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

	t.c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			// timeout return
			return
		default:
		}
		link := e.Attr("href")
		if link == "" {
			return
		}
		// Print link
		color.New(color.FgYellow).Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		err := t.c.Visit(e.Request.AbsoluteURL(link))
		fmt.Printf("%v\n", err)
	})
	// search for all link tags that have a rel attribute that is equal to stylesheet - CSS
	t.c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			// timeout return
			return
		default:
		}
		// hyperlink reference
		link := e.Attr("href")
		color.New(color.FgYellow).Println("Css found", "-->", link)
		// extraction
		t.Analyze(ctx, e.Request.AbsoluteURL(link))
	})

	// search for all script tags with src attribute -- JS
	t.c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			// timeout return
			return
		default:
		}
		// src attribute
		link := e.Attr("src")
		// Print link
		color.New(color.FgYellow).Println("Js found", "-->", link)
		// extraction
		t.Analyze(ctx, e.Request.AbsoluteURL(link))
	})

	// serach for all img tags with src attribute -- Images
	t.c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			// timeout return
			return
		default:
		}
		// src attribute
		link := e.Attr("src")
		if strings.HasPrefix(link, "data:image") || strings.HasPrefix(link, "blob:") {
			return
		}
		// Print link
		color.New(color.FgYellow).Println("img found", "-->", link)
		// extraction
		t.Analyze(ctx, e.Request.AbsoluteURL(link))
	})

	// Before making a request
	t.c.OnRequest(func(r *colly.Request) {
		select {
		case <-ctx.Done():
			// timeout return
			return
		default:
		}
		atomic.AddInt64(&t.urlCount, 1)
		link := r.URL.String()
		color.New(color.BgGreen).Println("try visit", "-->", link)
		if url == link {
			t.HtmlAnalyze(ctx, link)
		}
	})

	// Visit each url and wait for stuff to load :)
	if err := t.c.Visit(url); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	t.c.Wait()
}
