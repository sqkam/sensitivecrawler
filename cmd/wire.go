//go:build wireinject

package cmd

import (
	"github.com/google/wire"
	"github.com/sqkam/sensitivecrawler/ioc"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler"
)

func InitSensitiveCrawler() sensitivecrawler.Service {
	panic(wire.Build(ioc.InitConfig, ioc.InitSensitiveCrawler, ioc.InitSensitiveMatcher))
}
