package callbacker

import (
	"fmt"

	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"
)

type printCallBacker struct{}

func (c *printCallBacker) Do(ch <-chan result.Result) {
	for r := range ch {
		fmt.Printf("%v\n", r.Info)
	}
}

func NewPrintCallBacker() CallBacker {
	return &printCallBacker{}
}
