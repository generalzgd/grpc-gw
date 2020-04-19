/**
 * @version: 1.0.0
 * @author: zgd: general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: manager.go
 * @time: 2019-08-03 22:36
 */

package mgr

import (
	`fmt`
	`os`
	`path/filepath`
	`reflect`
	`testing`

	"github.com/generalzgd/comm-libs/conf/ymlcfg"
	"github.com/generalzgd/comm-libs/env"
	`github.com/generalzgd/grpc-svr-frame/common`
	`github.com/generalzgd/micro-cmdid/gocmd`
	u `github.com/generalzgd/micro-proto/goproto/user`
	`github.com/golang/protobuf/proto`

	`github.com/generalzgd/grpc-gw/config`
)

var (
	testCfg    config.AppConfig
	clientInfo = &common.ClientConnInfo{
		Uid:      163,
		Guid:     "12F44847-85EB-E6D8-6D38-09AA118BB528",
		Platform: 2,
	}
	p *Manager
)

func init() {
	pwd, _ := os.Getwd()
	pwd = filepath.Join(filepath.Dir(pwd), "config", fmt.Sprintf("config_%s.yml", env.GetEnvName()))

	testCfg.Load(pwd)
	//
	p = NewManager()
	p.Init(testCfg)
}

func TestManager_formatProtoStr(t *testing.T) {
	msg := &u.LoginRequest{
		Uid:      163,
		Token:    "34231",
		Platform: 1,
		TraceId:  "",
	}
	bts, _ := proto.Marshal(msg)

	msg.Token = ""
	bts2, _ := proto.Marshal(msg)

	type args struct {
		cmdid uint16
		body  []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "TestManager_formatProtoStr_1",
			args: args{
				cmdid: gocmd.IdLoginRequest,
				body:  bts,
			},
		},
		{
			name: "TestManager_formatProtoStr_2",
			args: args{
				cmdid: gocmd.IdLoginRequest,
				body:  bts2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.formatProtoStr(tt.args.cmdid, tt.args.body); got != tt.want {
				t.Logf("formatProtoStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_getEndpointByCmdId(t *testing.T) {
	type args struct {
		cmdId uint16
	}
	tests := []struct {
		name  string
		args  args
		want  ymlcfg.EndpointConfig
		want1 bool
	}{
		// TODO: Add test cases.
		{
			name: "TestManager_getEndpointByCmdId",
			args: args{cmdId: gocmd.IdLoginRequest},
			want: ymlcfg.EndpointConfig{
				Name:      "AuthSvr",
				Address:   "consul:///authsvr",
				Port:      0,
				Secure:    false,
				CertFiles: []ymlcfg.CertFile{{"cert.pem", "priv.key", "ca.pem"}},
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, got1 := p.getEndpointByCmdId(tt.args.cmdId)
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("getEndpointByCmdId() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Logf("getEndpointByCmdId() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
