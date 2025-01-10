package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/sqkam/sensitivecrawler/constant/features"
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
	depth                                int64
	timeout                              int64
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
		Args:  cobra.MinimumNArgs(1),
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
	rootCmd.PersistentFlags().Int64VarP(&depth, "depth", "", 0, "depth")
	rootCmd.PersistentFlags().Int64VarP(&timeout, "timeout", "t", 0, "timeout second")
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

var (
	maxMem  uint64     // 最大内存使用量(bytes)
	memLock sync.Mutex // 用于保护 maxMem 的互斥锁
)

// updateMaxMemory 更新最大内存使用量
func updateMaxMemory(memStats runtime.MemStats) {
	memLock.Lock()
	defer memLock.Unlock()
	if memStats.Alloc > maxMem {
		maxMem = memStats.Alloc
	}
}

// monitorMemory 持续监控内存使用
func monitorMemory() {
	var memStats runtime.MemStats
	for {
		runtime.ReadMemStats(&memStats)
		updateMaxMemory(memStats)
		time.Sleep(1 * time.Millisecond) // 控制监控频率，可以调整
	}
}

func run(cmd *cobra.Command, args []string) {
	go monitorMemory()
	if features.Debug {
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
		}()
	}

	startTime := time.Now()

	site = args[0]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := InitSensitiveCrawler()
	var options []sensitivecrawler.TaskOption

	options = append(options, sensitivecrawler.WithAllowedDomains(strings.Split(allowedDomains, ",")))

	if depth > 0 {
		options = append(options, sensitivecrawler.WithMaxDepth(int(depth)))
	}
	if timeout > 0 {
		options = append(options, sensitivecrawler.WithTimeOut(timeout))
	}
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
	options = append(options, sensitivecrawler.WithCallBacker(httpcallbacker.New(httpCallBackerUrl)))
	s.AddTask(site, options...)

	s.RunOneTask(ctx)

	if features.Debug {
		memLock.Lock()
		defer memLock.Unlock()
		maxMemMB := float64(maxMem) / float64(1024*1024) // 将 bytes 转换为 MB
		fmt.Printf("最大内存使用量: %.2f MB\n", maxMemMB)
		endTime := time.Now() // 记录结束时间

		duration := endTime.Sub(startTime) // 计算时间差

		durationInSeconds := duration.Seconds() // 将时间差转换为秒

		fmt.Printf("程序运行时间: %.2f 秒\n", durationInSeconds)
		time.Sleep(time.Second * 100)
	}
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
