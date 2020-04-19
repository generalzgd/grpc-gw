/**
 * @version: 1.0.0
 * @author: zhangguodong:general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: sessionhandler.go
 * @time: 2019/8/6 22:10
 */

package mgr

import (
	"testing"

	"github.com/generalzgd/grpc-svr-frame/common"
	lb `github.com/generalzgd/grpc-svr-frame/grpc-consul`
	"github.com/generalzgd/link"
)

// type Foo struct {
// 	A string
// }

var (
	/*clientInfo = &common.ClientConnInfo{
		Uid:      163,
		Guid:     "12F44847-85EB-E6D8-6D38-09AA118BB528",
		Platform: 2,
	}*/

	/*cfg = config.AppConfig{
		MemPool: ymlcfg.MemPoolConfig{
			Type:     "atom",
			Factor:   2,
			MinChunk: 64,
			MaxChunk: 65536,
			PageSize: 1048576,
		},
		EndpointSvr: map[string]ymlcfg.EndpointConfig{
			"authorizesvr": {
				Name:    "authorizesvr",
				Address: "consul:///authsvr",
			},
			"duoteamsvr": {
				Name:    "duoteamsvr",
				Address: "consul:///duoteamsvr",
			},
			"hfbrosvr": {
				Name:    "hfbrosvr",
				Address: "consul:///hfbrosvr",
			},
		},
	}*/
)

// func Test_getStructName(t *testing.T) {
// 	type args struct {
// 		instance interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			name: "Test_getStructName",
// 			args: args{
// 				instance: &Foo{},
// 			},
// 			want: "Foo",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := getStructName(tt.args.instance); got != tt.want {
// 				t.Errorf("getStructName() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }






func TestManager_onUserLogin(t *testing.T) {
	lb.InitRegister("http://127.0.0.1:8500")

	p := NewManager()
	p.Init(testCfg)
	type args struct {
		info    *common.ClientConnInfo
		session *link.Session
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "TestManager_onUserLogin",
			args: args{
				info:    &common.ClientConnInfo{},
				session: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.onUserLogin(tt.args.info, tt.args.session)
		})
	}
}

//func TestManager_handleLogin(t *testing.T) {
//	data := &user.HfLoginRequest{
//		Uid:      195,
//		Token:    "rn5qh83ps58j7l4c9216htd2p9",
//		Platform: 128,
//		TraceId:  "1234",
//	}
//	bts, err := proto.Marshal(data)
//	if err != nil {
//		return
//	}
//
//	pack := gatepack.GateClientPack{
//		GateClientPackHead: gatepack.GateClientPackHead{
//			Length: uint16(len(bts)),
//			Seq:    0,
//			Cmdid:  8193,
//			Ver:    100,
//			Codec:  0,
//			Opt:    0,
//		},
//		Body: bts,
//	}
//
//	type args struct {
//		session *link.Session
//		pack    gatepack.GateClientPack
//		info    *common.ClientConnInfo
//	}
//	tests := []struct {
//		name   string
//		args   args
//		want   bool
//		want1  zqerr.IErrCode
//	}{
//		// TODO: Add test cases.
//		{
//			name:"TestManager_handleLogin",
//			args:args{
//				session: nil,
//				pack:    pack,
//				info:    &common.ClientConnInfo{
//					State:     false,
//				},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, got1 := p.handleLogin(tt.args.session, tt.args.pack, tt.args.info)
//			if got != tt.want {
//				t.Errorf("handleLogin() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("handleLogin() got1 = %v, want %v", got1, tt.want1)
//			}
//		})
//	}
//}