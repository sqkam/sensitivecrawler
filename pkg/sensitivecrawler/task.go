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

type task struct {
	site       string
	callBacker callbacker.CallBacker
	m          sensitivematcher.SensitiveMatcher
	callerCh   chan result.Result
	c          *colly.Collector
	urlCount   int64
}

func (t *task) Analyze(url string) {
	fmt.Println("Analyzing --> ", url)
	// get the html body
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return
	}
	// Closure
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err == nil {
		matchStr, ok := t.m.Match(respBody, url)
		if ok {
			t.callerCh <- result.Result{Site: t.site, Url: url, Info: matchStr}
		}

	}
}

func (t *task) HtmlAnalyze(url string) {
	fmt.Println("Analyzing --> ", url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// get the html body
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return
	}
	// Closure
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err == nil {
		matchStr, ok := t.m.Match(respBody, url)
		if ok {
			t.callerCh <- result.Result{Url: url, Info: matchStr}
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
		t.Analyze(e.Request.AbsoluteURL(link))
	})

	// search for all script tags with src attribute -- JS
	t.c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		// src attribute
		link := e.Attr("src")
		// Print link
		// fmt.Println("Js found", "-->", link)
		// extraction
		t.Analyze(e.Request.AbsoluteURL(link))
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
		t.Analyze(e.Request.AbsoluteURL(link))
	})

	// Before making a request
	t.c.OnRequest(func(r *colly.Request) {
		atomic.AddInt64(&t.urlCount, 1)
		link := r.URL.String()
		if url == link {
			t.HtmlAnalyze(link)
		}
	})

	// Visit each url and wait for stuff to load :)
	if err := t.c.Visit(url); err != nil {
		return
	}

	t.c.Wait()
}
