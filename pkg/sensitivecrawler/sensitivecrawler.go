package sensitivecrawler

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/imthaghost/goclone/pkg/parser"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivematcher"
	"io"
	"net/http"
	"net/http/cookiejar"

	"strings"
)

type service struct {
	m sensitivematcher.SensitiveMatcher
}

func (s *service) Analyze(url string) {

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
		s.m.Match(respBody, url)
	}

}

func (s *service) HtmlAnalyze(url string) {

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
		s.m.Match(respBody, url)
	}

}

func (s *service) Run(ctx context.Context) {
	url := "http://vcrm.4paradigm.com"

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

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}
	c := colly.NewCollector(
		colly.Async(true),
	)
	c.SetCookieJar(jar)

	// search for all link tags that have a rel attribute that is equal to stylesheet - CSS
	c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		// hyperlink reference
		link := e.Attr("href")
		// print css file was found
		fmt.Println("Css found", "-->", link)
		// extraction
		s.Analyze(e.Request.AbsoluteURL(link))
	})

	// search for all script tags with src attribute -- JS
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		// src attribute
		link := e.Attr("src")
		// Print link
		fmt.Println("Js found", "-->", link)
		// extraction
		s.Analyze(e.Request.AbsoluteURL(link))
	})

	// serach for all img tags with src attribute -- Images
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		// src attribute
		link := e.Attr("src")
		if strings.HasPrefix(link, "data:image") || strings.HasPrefix(link, "blob:") {
			return
		}
		// Print link
		fmt.Println("Img found", "-->", link)
		// extraction
		s.Analyze(e.Request.AbsoluteURL(link))
	})

	//Before making a request
	c.OnRequest(func(r *colly.Request) {
		link := r.URL.String()
		if url == link {
			s.HtmlAnalyze(link)
		}
	})

	// Visit each url and wait for stuff to load :)
	if err := c.Visit(url); err != nil {
		return
	}

	c.Wait()

}

func NewDefaultService(m sensitivematcher.SensitiveMatcher) Service {
	return &service{
		m: m,
	}
}
