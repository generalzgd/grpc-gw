/**
 * @version: 1.0.0
 * @author: generalzgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: main.go
 * @time: 2020/3/3 11:43 上午
 * @project: hfgateway
 */

package main

import (
	`fmt`
	`net/http`
	`os`
	"os/signal"
	`runtime`
	"syscall"

	`github.com/astaxie/beego/logs`
	`github.com/generalzgd/comm-libs/consul`
	`github.com/generalzgd/comm-libs/file`

	`github.com/generalzgd/grpc-gw/config`
	`github.com/generalzgd/grpc-gw/mgr`
)

func init() {
	file.DetectDir(file.LogsDir)

	logger := logs.GetBeeLogger()
	logger.SetLevel(logs.LevelInfo)
	logger.SetLogger(logs.AdapterConsole)
	logger.SetLogger(logs.AdapterFile, `{"filename":"logs/file.log","level":7,"maxlines":1024000000,"maxsize":1024000000,"daily":true,"maxdays":7}`)
	logger.EnableFuncCallDepth(true)
	logger.SetLogFuncCallDepth(3)
	logger.Async(100000)
}

func exit(err error) {
	code := 0
	if err != nil {
		logs.Error("exit() got err=[%v]", err)
		code = 1
	}
	logs.GetBeeLogger().Flush()
	os.Exit(code)
}

func main() {
	var err error
	defer func() {
		exit(err)
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg := config.AppConfig{}
	if err = cfg.Load(""); err != nil {
		return
	}

	logs.SetLevel(cfg.GetLogLevel())

	if cfg.Pprof > 0 {
		go http.ListenAndServe(fmt.Sprintf(":%d", cfg.Pprof), nil)
	}

	manager := mgr.NewManager()
	manager.Init(cfg)

	if err = manager.ServeClient(); err != nil {
		return
	}

	if err = manager.ServeGrpc(); err != nil {
		return
	}

	// 注册服务
	agent := consul.GetConsulRemoteInst(cfg.Consul.Address, cfg.Consul.Token)
	agent.Init(cfg.Name, "grpc", cfg.RpcSvr.Port, 0, cfg.Consul.HealthPort, cfg.Consul.HealthType, "")
	if err = agent.Register(nil, nil); err == nil {
		logs.Info("Consul register success.")
	} else {
		logs.Error("Consul register fail.", err)
	}

	// catch system signal
	sig := watchSignal()
	logs.Info("signal:", sig)

	// 注销服务
	if err = agent.Deregister(""); err == nil {
		logs.Info("Consul deregister success.")
	} else {
		logs.Error("Consul deregister fail.", err)
	}
	agent.Destroy()
	manager.Destroy()
}

// 监控系统的信号，会阻塞当前协程
func watchSignal() os.Signal {
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTERM)
	sig := <-chSig
	return sig
}