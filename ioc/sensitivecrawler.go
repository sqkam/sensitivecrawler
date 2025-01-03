package ioc

import (
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivematcher"
)

func InitSensitiveCrawler(m sensitivematcher.SensitiveMatcher) sensitivecrawler.Service {
	return sensitivecrawler.NewDefaultService(m)
}
