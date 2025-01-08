package cmd

import (
	"context"
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
		Run:   run,
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

func run(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := InitSensitiveCrawler()
	s.AddTask("https://zgo.sqkam.cfd")

	s.RunOneTask(ctx)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//func main() {
//	quit := make(chan os.Signal)
//	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	s := InitSensitiveCrawler()
//	go func() {
//		s.Run(ctx)
//	}()
//	<-quit
//}
