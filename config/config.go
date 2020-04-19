/**
 * @version: 1.0.0
 * @author: generalzgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: config.go
 * @time: 2020/3/3 2:36 下午
 * @project: hfgateway
 */

package config

import (
	`fmt`
	`os`
	`path/filepath`
	`time`

	`github.com/astaxie/beego/logs`
	`github.com/generalzgd/comm-libs/conf/ymlcfg`
	"github.com/generalzgd/comm-libs/env"
)

type GateConnConfig struct {
	Name            string            `yaml:"name"`
	Type            string            `yaml:"type"`
	Secure          bool              `yaml:"secure"`
	CertFiles       []ymlcfg.CertFile `yaml:"certfiles"`
	BufferSize      int               `yaml:"buffersize"`
	MaxConn         int               `yaml:"maxconn"`
	IdleTimeout     time.Duration     `yaml:"idletimeout"`
	SendChanSize    int               `yaml:"sendchansize"`
	ReceiveChanSize int               `yaml:"recvchansize"`
	Port            int               `yaml:"port"`
}

type AppConfig struct {
	Name          string                           `yaml:"name"`
	Ver           string                           `yaml:"ver"`
	LogLevel      int                              `yaml:"loglevel"`
	Pprof         int                              `yaml:"pprof"`
	MailAddr      []string                         `yaml:"mailaddr"`
	Consul        ymlcfg.ConsulConfig              `yaml:"consul"`
	MemPool       ymlcfg.MemPoolConfig             `yaml:"mempool"`
	EndpointSvr   map[string]ymlcfg.EndpointConfig `yaml:"endpoint"`
	RpcSvr        ymlcfg.EndpointConfig            `yaml:"rpc"`
	ServeList     []GateConnConfig                 `yaml:"servelist"`
	HealthCheckIp []string                         `yaml:"healthcheckip"`
}

func (p *AppConfig) GetLogLevel() int {
	if p.LogLevel <= 0 {
		return logs.LevelInfo
	}
	return p.LogLevel
}

func (p *AppConfig) Load(path string) error {
	if len(path) < 1 {
		path = fmt.Sprintf("%s/config/config_%s.yml", filepath.Dir(os.Args[0]), env.GetEnvName())
	}
	if err := ymlcfg.LoadYaml(path, p); err != nil {
		return err
	}
	return nil
}
