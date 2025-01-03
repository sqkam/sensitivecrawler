package main

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/sqkam/sensitivecrawler/config"
)

func TestAsdafsadf(t *testing.T) {
	var conf config.Config
	v := viper.New()
	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&conf); err != nil {
		panic(err)
	}
}
