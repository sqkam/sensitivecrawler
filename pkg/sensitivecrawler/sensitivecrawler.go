package sensitivecrawler

import (
	"context"
	"fmt"
	"net/url"
	"time"

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
	})
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
	_ = u

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
