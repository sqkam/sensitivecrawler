package config

type Rule struct {
	Name     string `mapstructure:"name"`
	Exp      string `mapstructure:"exp"`
	GroupIdx int    `mapstructure:"group_idx"`
}

type Config struct {
	ParallelCount int64  `mapstructure:"parallel_count"`
	Rules         []Rule `mapstructure:"rules"`
}
