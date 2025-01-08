package ioc

import (
	"github.com/spf13/viper"
	"github.com/sqkam/sensitivecrawler/config"
)

func InitConfig() config.Config {
	var conf config.Config
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&conf); err != nil {
		panic(err)
	}

	return conf
}
