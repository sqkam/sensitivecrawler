package callbacker

import "github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/result"

type CallBacker interface {
	Do(ch <-chan result.Result)
}
