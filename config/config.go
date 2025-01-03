package config

type Rule struct {
	Name string `mapstructure:"name"`
	Exp  string `mapstructure:"exp"`
}

type Config struct {
	Rules []Rule `mapstructure:"rules"`
}
