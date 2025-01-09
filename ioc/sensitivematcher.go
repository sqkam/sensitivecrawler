package ioc

import (
	"github.com/sqkam/sensitivecrawler/config"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivematcher"
)

func InitSensitiveMatcher(c config.Config) sensitivematcher.SensitiveMatcher {
	return sensitivematcher.NewAsyncSensitiveMatcher(c)
}
