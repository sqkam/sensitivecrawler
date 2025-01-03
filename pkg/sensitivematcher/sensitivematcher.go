package sensitivematcher

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/sqkam/sensitivecrawler/config"
	"golang.org/x/sync/errgroup"
)
import "github.com/fatih/color"

var greenWriter = color.New(color.FgGreen).SprintFunc()
var yellowWriter = color.New(color.FgYellow).SprintFunc()
var redWriter = color.New(color.FgRed).SprintFunc()
var defaultWriter = color.New().SprintFunc()

type sensitiveMatcher struct {
	rules []config.Rule
}

func (m *sensitiveMatcher) Match(b []byte, name string) {
	var eg errgroup.Group
	for _, v := range m.rules {
		v := v
		eg.Go(func() error {
			exp := regexp.MustCompile(v.Exp, regexp.None)
			match, err := exp.FindStringMatch(string(b))
			if err != nil {
				return nil
			}
			if match != nil && match.GroupCount() > 1 {
				redWriter(name, " 发现敏感信息 ", v.Name, ": ", match.Groups()[1].String())
			}
			return nil
		})
	}
	eg.Wait()
}

func NewDefaultMatcher(c config.Config) SensitiveMatcher {
	return &sensitiveMatcher{
		rules: c.Rules,
	}
}
