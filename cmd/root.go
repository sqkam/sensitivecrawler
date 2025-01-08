package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler"
	httpcallbacker "github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker/http"
	retryablehttpcallbacker "github.com/sqkam/sensitivecrawler/pkg/sensitivecrawler/callbacker/retryablehttp"

	"github.com/spf13/cobra"
)

const (
	appLogo = `
                                      ███   █████     ███                                                                    ████                    
                                     ░░░   ░░███     ░░░                                                                    ░░███                    
  █████   ██████  ████████    █████  ████  ███████   ████  █████ █████  ██████   ██████  ████████   ██████   █████ ███ █████ ░███   ██████  ████████ 
 ███░░   ███░░███░░███░░███  ███░░  ░░███ ░░░███░   ░░███ ░░███ ░░███  ███░░███ ███░░███░░███░░███ ░░░░░███ ░░███ ░███░░███  ░███  ███░░███░░███░░███
░░█████ ░███████  ░███ ░███ ░░█████  ░███   ░███     ░███  ░███  ░███ ░███████ ░███ ░░░  ░███ ░░░   ███████  ░███ ░███ ░███  ░███ ░███████  ░███ ░░░ 
 ░░░░███░███░░░   ░███ ░███  ░░░░███ ░███   ░███ ███ ░███  ░░███ ███  ░███░░░  ░███  ███ ░███      ███░░███  ░░███████████   ░███ ░███░░░   ░███     
 ██████ ░░██████  ████ █████ ██████  █████  ░░█████  █████  ░░█████   ░░██████ ░░██████  █████    ░░████████  ░░████░████    █████░░██████  █████    
░░░░░░   ░░░░░░  ░░░░ ░░░░░ ░░░░░░  ░░░░░    ░░░░░  ░░░░░    ░░░░░     ░░░░░░   ░░░░░░  ░░░░░      ░░░░░░░░    ░░░░ ░░░░    ░░░░░  ░░░░░░  ░░░░░
`
	appDesc = "a powerful, lightning and fast sensitive information detection tools"
)

var appAboutLong = fmt.Sprintf("%s\n%s", appLogo, appDesc)

var (
	cfgFile                              string
	allowedDomains                       string
	site                                 string
	httpCallBackerUrl                    string
	retryableHttpCallBackerUrl           string
	retryableHttpCallBackerRetryCount    int64
	retryableHttpCallBackerRetryInterval int64
	rootCmd                              = &cobra.Command{
		Use:   "sensitivecrawler site",
		Short: appDesc,
		Long:  appAboutLong,
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
	rootCmd.PersistentFlags().StringVarP(&allowedDomains, "allowedDomains", "d", "", "allowedDomains separated by commas")
	rootCmd.PersistentFlags().StringVarP(&allowedDomains, "depth", "", "", "allowedDomains separated by commas")
	rootCmd.PersistentFlags().StringVarP(&httpCallBackerUrl, "httpCallBackerUrl", "", "", "httpCallBackerUrl")
	rootCmd.PersistentFlags().StringVarP(&retryableHttpCallBackerUrl, "retryableHttpCallBackerUrl", "", "", "retryableHttpCallBackerUrl")
	rootCmd.PersistentFlags().Int64VarP(&retryableHttpCallBackerRetryCount, "retryableHttpCallBackerRetryCount", "", 3, "set retryableHttpCallBackerRetryCount second")
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
	if len(args) < 1 {
		panic(errors.New("please input site"))
	}
	site = args[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := InitSensitiveCrawler()
	options := []sensitivecrawler.TaskOption{}

	options = append(options, sensitivecrawler.WithAllowedDomains(strings.Split(allowedDomains, ",")))
	options = append(options, sensitivecrawler.WithMaxDepth(1))
	if httpCallBackerUrl != "" {
		options = append(options, sensitivecrawler.WithCallBacker(httpcallbacker.New(httpCallBackerUrl)))
	}
	if retryableHttpCallBackerUrl != "" {
		options = append(options, sensitivecrawler.WithCallBacker(retryablehttpcallbacker.New(
			retryableHttpCallBackerUrl,
			retryablehttpcallbacker.WithRetryMax(retryableHttpCallBackerRetryCount),
			retryablehttpcallbacker.WithRetryInterval(time.Duration(retryableHttpCallBackerRetryInterval)*time.Second),
		)))
	}

	s.AddTask(site, options...)

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
