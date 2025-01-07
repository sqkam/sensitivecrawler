package sensitivecrawler

import (
	"context"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	print2 "github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker/print"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivematcher"
)

type service struct {
	taskCh        chan *task
	parallelCount int64
	taskCount     int64
	m             sensitivematcher.SensitiveMatcher
}

func (s *service) runTask(ctx context.Context, t *task) {
	if t.callBacker != nil {
		t.callBacker.Do(t.resultMsgCh)
	}
	var cancel context.CancelFunc
	if t.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(t.timeout)*time.Second)
		defer cancel()
	}
	t.Run(ctx)
	// 统计信息
	t.resultMsgCh <- result.Result{Site: t.site, Url: "", Info: "", Statistics: &result.Statistics{
		UrlCount: t.urlCount,
		// MemoryTotal:
	}}
	close(t.resultMsgCh)
}

func (s *service) AddTask(site string, options ...TaskOption) {
	c := colly.NewCollector(
		colly.Async(true),
	)
	extensions.RandomUserAgent(c)
	u, err := url.Parse(site)
	if err != nil {
		return
	}
	c.AllowedDomains = []string{u.Hostname()}

	t := &task{
		site:        site,
		m:           s.m,
		callBacker:  print2.NewPrintCallBacker(),
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

func NewDefaultService(m sensitivematcher.SensitiveMatcher) Service {
	return &service{
		m: m,
	}
}
