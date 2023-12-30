package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"pp/config"
	"pp/log"
	"pp/service"
	"runtime"
	"syscall"
	"time"
)

func runPProfServer() {
	_ = http.ListenAndServe("0.0.0.0:6061", nil)
}

func main() {
	go runPProfServer()
	// 加载app.json需要第一时间启动
	appConfigLoad := config.NewAppConfig()
	appConfigLoad.LoadConfig()

	// 初始化日志系统
	logger := log.GetLogger()
	if !logger.InitLogger() {
		fmt.Println("init logger error")
		return
	}

	// 服务器启动相关初始化
	svrLibHandler := service.GetSvrlibhandler()
	if !svrLibHandler.OnInit() {
		logger.Error("svrLibHandler.OnInit failed")
		return
	}

	// 启动玩家消息处理
	go service.StartMessageProcess()
	// 获取app.json配置
	appConfig := appConfigLoad.GetConfig()
	logger.Info("Start server:" + appConfig.ServerName)

	go func() {
		for {
			logger.Info("current Goroutine count:", runtime.NumGoroutine())
			time.Sleep(time.Second)
		}
	}()

	c := make(chan os.Signal, 1) // ---> 优雅重启
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		logger.Info("get a signal ", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			svrLibHandler.OnQuit()
			return
		case syscall.SIGHUP:
			svrLibHandler.ReloadAppConfig()
			logger.Info("reload app.json success")
		default:
			return
		}
	}
}
