package print

import (
	"fmt"

	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"
)

type printCallBacker struct{}

func (c *printCallBacker) Do(ch <-chan result.Result) {
	for r := range ch {
		go c.doCallback(r)
	}
}

func (c *printCallBacker) doCallback(r result.Result) {
	fmt.Printf("%v\n", r.Info)
}

func NewPrintCallBacker() callbacker.CallBacker {
	return &printCallBacker{}
}
