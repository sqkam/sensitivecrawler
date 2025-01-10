package sensitivematcher

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	regexp "github.com/dlclark/regexp2"
	"github.com/sqkam/sensitivecrawler/config"
)

type asyncSensitiveMatcher struct {
	rules []config.Rule
	exps  []*regexp.Regexp
}

func (m *asyncSensitiveMatcher) Match(ctx context.Context, b []byte) []string {
	var eg errgroup.Group
	var result []string

	strCh := make(chan string, 10)
	waitStrCh := make(chan struct{}, 1)
	go func() {
		for v := range strCh {
			result = append(result, v)
		}
		waitStrCh <- struct{}{}
	}()
	for i, v := range m.rules {
		v := v
		i := i
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:

			}
			exp := m.exps[i]
			rb := bytesToRunes(b)
			defer func() {
				rb = nil
			}()
			match, err := exp.FindRunesMatch(rb)
			if err != nil {
				return nil
			}
			if match != nil && match.GroupCount() > v.GroupIdx {
				strCh <- fmt.Sprintf("发现敏感信息 %s: %s", v.Name, match.Groups()[v.GroupIdx].String())
			}
			return nil
		})
	}
	_ = eg.Wait()
	close(strCh)
	<-waitStrCh
	close(waitStrCh)
	return result
}

func NewAsyncSensitiveMatcher(c config.Config) SensitiveMatcher {
	s := &asyncSensitiveMatcher{
		rules: c.Rules,
	}
	exps := make([]*regexp.Regexp, len(s.rules))
	for i, v := range s.rules {
		exp := regexp.MustCompile(v.Exp, regexp.None)
		exps[i] = exp
	}
	s.exps = exps
	return s
}
