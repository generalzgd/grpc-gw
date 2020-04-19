/**
 * @version: 1.0.0
 * @author: zgd: general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: imgatedown_handler.go
 * @time: 2019-08-12 20:14
 */

package mgr

import (
	"context"
	`fmt`
	"net"
	"reflect"
	"testing"

	"github.com/generalzgd/grpc-svr-frame/common"
	"github.com/generalzgd/link"
	"github.com/generalzgd/micro-cmdid/gocmd"
	`github.com/generalzgd/micro-proto/goproto/comm`
	`github.com/generalzgd/micro-proto/goproto/core`
	u `github.com/generalzgd/micro-proto/goproto/user`
	"github.com/golang/protobuf/proto"
)

func Test_Main(t *testing.T) {
	fmt.Printf("<< %X", 2542)
}

func TestManager_TransmitProtoDown(t *testing.T) {


	tcpcodec, _ := newTcpCodec(&net.TCPConn{}, 1000, &p.GateClientPackEncrypt)
	p.linkCenter.Put(163, 1, link.NewSession(tcpcodec, 100))
	p.linkCenter.Put(163, 2, link.NewSession(tcpcodec, 100))

	type args struct {
		info  *common.ClientConnInfo
		msg   proto.Message
		cmdid uint16
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "TestManager_TransmitProtoDown",
			args: args{
				info: clientInfo,
				msg: &u.GetUserInfoReply{
					Code:    0,
					Message: "ok",
					User: []*u.UserInfo{
						{
							Uid:      163,
							Nickname: "你好我好大家好",
							Avatar:   "sdfsdf",
							Gender:   1,
						},
					},
				},
				cmdid: gocmd.IdGetUserInfoRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.TransmitProtoDown(tt.args.info, tt.args.msg, tt.args.cmdid)
		})
	}
}

func TestManager_OnHfLogin(t *testing.T) {

	type args struct {
		ctx context.Context
		req *u.LoginNotify
	}
	tests := []struct {
		name      string
		args      args
		wantReply *comm.CommReply
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "TestManager_OnLogin",
			args: args{
				ctx: p.MakeIncomingContextByClientInfo(context.Background(), clientInfo),
				req: &u.LoginNotify{
					User: &u.OneUser{
						Uid:      163,
						Platform: 2,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReply, err := p.OnLogin(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Logf("Manager.OnLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("got reply:%v", gotReply)
		})
	}
}

func TestManager_Kickout(t *testing.T) {

	type args struct {
		ctx context.Context
		req *u.KickOutRequest
	}
	tests := []struct {
		name      string
		args      args
		wantReply *comm.Error
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "TestManager_Kickout",
			args: args{
				ctx: p.MakeIncomingContextByClientInfo(context.Background(), clientInfo),
				req: &u.KickOutRequest{
					Uid:          163,
					LastPlatform: 2,
					LastSid:      1,
				},
			},
			wantReply: &comm.Error{},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReply, err := p.KickOut(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Logf("Manager.Kickout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Logf("Manager.Kickout() = %v, want %v", gotReply, tt.wantReply)
			}
		})
	}
}

func TestManager_Notify(t *testing.T) {
	note := &core.TeamInfo{
	}
	bts, _ := proto.Marshal(note)

	type args struct {
		ctx context.Context
		req *u.MsgNotify
	}
	tests := []struct {
		name    string
		args    args
		want    *comm.CommReply
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "TestManager_Notify",
			args: args{
				ctx: p.MakeIncomingContextByClientInfo(context.Background(), clientInfo),
				req: &u.MsgNotify{
					User:    []*u.OneUser{},
					CmdId:   100,// gocmd.ID_DuoTeamInfo,
					Message: bts,
				},
			},
			want:&comm.CommReply{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Notify(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Logf("Manager.Notify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("Manager.Notify() = %v, want %v", got, tt.want)
			}
		})
	}
}
