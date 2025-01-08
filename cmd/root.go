package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler"
	httpcallbacker "github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker/http"
	retryablehttpcallbacker "github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker/retryablehttp"

	"github.com/spf13/cobra"
)

const (
	appDesc = "sensitivecrawler"
)

var (
	cfgFile                              string
	site                                 string
	httpCallBackerUrl                    string
	retryableHttpCallBackerUrl           string
	retryableHttpCallBackerRetryCount    int64
	retryableHttpCallBackerRetryInterval int64
	rootCmd                              = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVarP(&site, "site", "s", "", "site to scan")
	rootCmd.PersistentFlags().StringVarP(&httpCallBackerUrl, "httpCallBackerUrl", "", "", "httpCallBackerUrl")
	rootCmd.PersistentFlags().StringVarP(&retryableHttpCallBackerUrl, "retryableHttpCallBackerUrl", "", "", "retryableHttpCallBackerUrl")
	rootCmd.PersistentFlags().Int64VarP(&retryableHttpCallBackerRetryCount, "retryableHttpCallBackerRetryCount", "", 3, " set retryableHttpCallBackerRetryCount second")
	rootCmd.PersistentFlags().Int64VarP(&retryableHttpCallBackerRetryInterval, "retryableHttpCallBackerRetryInterval", "", 3, "retryableHttpCallBackerRetryInterval")
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
	options := []sensitivecrawler.TaskOption{}
	if httpCallBackerUrl != "" {
		options = append(options, sensitivecrawler.WithCallBacker(httpcallbacker.New(httpCallBackerUrl)))
	}
	if retryableHttpCallBackerUrl != "" {
		options = append(options, sensitivecrawler.WithCallBacker(retryablehttpcallbacker.New(
			httpCallBackerUrl,
			retryablehttpcallbacker.WithRetryMax(retryableHttpCallBackerRetryCount),
			retryablehttpcallbacker.WithRetryInterval(time.Duration(retryableHttpCallBackerRetryInterval)*time.Second),
		)))
	}
	if site == "" {
		panic(errors.New("please input site"))
	}
	s.AddTask(site)

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
