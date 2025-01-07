package sensitivecrawler

import (
	"context"
	"net/http/cookiejar"

	"github.com/gocolly/colly/v2"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker"
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
		t.callBacker.Do(t.callerCh)
	}
	t.Run(ctx)
	// 统计信息
	t.callerCh <- result.Result{Site: t.site, Url: "", Info: "", Statistics: &result.Statistics{
		UrlCount: t.urlCount,
		// MemoryTotal:
	}}
	close(t.callerCh)
}

func (s *service) AddTask(url string, callBacker callbacker.CallBacker) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}
	c := colly.NewCollector(
		colly.Async(true),
	)
	c.SetCookieJar(jar)
	s.taskCh <- &task{
		site:       url,
		callBacker: callBacker,
		m:          s.m,
		callerCh:   make(chan result.Result, 30),
		c:          c,
	}
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
