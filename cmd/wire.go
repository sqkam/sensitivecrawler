//go:build wireinject

package cmd

import (
	"github.com/google/wire"
	"github.com/sqkam/sensitivecrawler/ioc"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivematcher"
)

func InitSensitiveCrawler() sensitivecrawler.Service {
	panic(wire.Build(ioc.InitConfig, sensitivecrawler.NewDefaultService, sensitivematcher.NewDefaultMatcher))
}
