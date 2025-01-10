package sensitivematcher

import (
	"context"
	"fmt"

	regexp "github.com/dlclark/regexp2"
	"github.com/sqkam/sensitivecrawler/config"
)
import "github.com/fatih/color"

var (
	greenWriter   = color.New(color.FgGreen).SprintFunc()
	yellowWriter  = color.New(color.FgYellow).SprintFunc()
	redWriter     = color.New(color.FgRed).SprintFunc()
	defaultWriter = color.New().SprintFunc()
)

type sensitiveMatcher struct {
	rules []config.Rule
	exps  []*regexp.Regexp
}

func (m *sensitiveMatcher) Match(ctx context.Context, b []byte) []string {
	var result []string

	strCh := make(chan string, 10)

	go func() {
		for i, v := range m.rules {
			select {
			case <-ctx.Done():
				continue
			default:
			}
			exp := m.exps[i]
			rb := bytesToRunes(b)
			match, err := exp.FindRunesMatch(rb)
			rb = nil
			if err != nil {
				continue
			}
			if match != nil && match.GroupCount() > v.GroupIdx {
				strCh <- fmt.Sprintf("发现敏感信息 %s: %s", v.Name, match.Groups()[v.GroupIdx].String())
			}
		}
		close(strCh)
	}()

	for v := range strCh {
		result = append(result, v)
	}
	return result
}

func NewDefaultMatcher(c config.Config) SensitiveMatcher {
	s := &sensitiveMatcher{
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
