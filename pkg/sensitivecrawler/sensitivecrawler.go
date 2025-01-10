package sensitivecrawler

import (
	"context"
	"fmt"
	"time"

	"github.com/imthaghost/goclone/pkg/parser"

	printcallbacker "github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker/print"

	"github.com/sqkam/sensitivecrawler/config"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"

	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivematcher"
)

type service struct {
	parallelCount  int64
	m              sensitivematcher.SensitiveMatcher
	totalTaskCount int64
	taskCh         chan *task
}

func NewDefaultService(c config.Config, m sensitivematcher.SensitiveMatcher) Service {
	return &service{
		parallelCount: c.ParallelCount,
		m:             m,
		taskCh:        make(chan *task, 10),
	}
}

func (s *service) runTask(ctx context.Context, t *task) {
	if t.callBacker != nil {
		go func() {
			t.callBacker.Do(t.resultMsgCh)
		}()
	}
	var cancel context.CancelFunc
	if t.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(t.timeout)*time.Second)
		defer cancel()
	}
	t.Run(ctx)
	// 统计信息
	t.resultMsgCh <- result.Result{Site: t.site, Url: "", Info: "", Statistics: &result.Statistics{
		UrlCount:       t.urlCount,
		SensitiveCount: t.sensitiveCount,
		// MemoryTotal:
	}}

	fmt.Printf("统计信息%#v\n", &result.Statistics{
		UrlCount:       t.urlCount,
		SensitiveCount: t.sensitiveCount,
		AnalyzeCount:   t.analyzeCount,
		AnalyzeBytes:   t.analyzeBytes / (1024 * 1024),
	})
	close(t.resultMsgCh)
}

// NewTask
func (s *service) AddTask(site string, options ...TaskOption) {
	// url := "https://zgo.sqkam.cfd"
	isValid, isValidDomain := parser.ValidateURL(site), parser.ValidateDomain(site)
	if !isValid && !isValidDomain {
		fmt.Printf("%q is not valid", site)
		return
	}

	domain := site
	if isValidDomain {
		site = parser.CreateURL(domain)
	} else {
		domain = parser.GetDomain(site)
	}
	c := colly.NewCollector(
		colly.Async(true),
	)
	//c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})
	extensions.RandomUserAgent(c)

	c.AllowedDomains = []string{domain}

	t := &task{
		site:        site,
		m:           s.m,
		callBacker:  printcallbacker.New(),
		resultMsgCh: make(chan result.Result, 30),
		c:           c,
	}
	for _, o := range options {
		o.Apply(t)
	}
	s.taskCh <- t
}

func (s *service) Run(ctx context.Context) {
	for range s.parallelCount {
		go func() {
			for v := range s.taskCh {
				s.runTask(ctx, v)
			}
		}()
	}
}

func (s *service) RunOneTask(ctx context.Context) {
	v := <-s.taskCh
	s.runTask(ctx, v)
}
