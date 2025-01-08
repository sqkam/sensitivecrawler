package ioc

import (
	"github.com/spf13/viper"
	"github.com/sqkam/sensitivecrawler/config"
)

func InitConfig() config.Config {
	// default config
	conf := config.Config{
		ParallelCount: 5,
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&conf); err != nil {
		panic(err)
	}

	return conf
}
