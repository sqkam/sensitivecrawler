package sensitivematcher

import (
	"fmt"

	regexp "github.com/dlclark/regexp2"
	"github.com/sqkam/sensitivecrawler/config"
	"golang.org/x/sync/errgroup"
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
}

func (m *sensitiveMatcher) Match(b []byte, name string) (string, bool) {
	var eg errgroup.Group
	var result string
	var matchAny bool
	strCh := make(chan string, 10)
	waitStrCh := make(chan struct{}, 1)
	go func() {
		for v := range strCh {
			result = v
		}
		waitStrCh <- struct{}{}
	}()
	for _, v := range m.rules {
		v := v
		eg.Go(func() error {
			exp := regexp.MustCompile(v.Exp, regexp.None)
			match, err := exp.FindStringMatch(string(b))
			if err != nil {
				return nil
			}
			if match != nil && match.GroupCount() > v.GroupIdx {
				strCh <- fmt.Sprintln(name, " 发现敏感信息 ", v.Name, ": ", match.Groups()[v.GroupIdx].String())
				matchAny = true
			}
			return nil
		})
	}
	eg.Wait()
	close(strCh)
	<-waitStrCh
	close(waitStrCh)
	return result, matchAny
}

func NewDefaultMatcher(c config.Config) SensitiveMatcher {
	return &sensitiveMatcher{
		rules: c.Rules,
	}
}
