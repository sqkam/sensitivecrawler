package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

const (
	appDesc = "sensitivecrawler"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "sensitivecrawler",
		Short: appDesc,
		Long:  appDesc + "long",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%v\n", "asdfasdf")
		},
	}
)

func init() {
	initFlags()
	cobra.OnInitialize(initConfig)

	// cobra.OnInitialize(initLogger)
}

func initFlags() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "conf", "c", "", "path of config file")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.SupportedExts = append([]string{"yaml", "yml"}, viper.SupportedExts...)
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.sensitivecrawler")
		viper.AddConfigPath("/etc/sensitivecrawler/")
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
