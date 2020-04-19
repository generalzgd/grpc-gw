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
	"fmt"

	"github.com/astaxie/beego/logs"
	libs `github.com/generalzgd/comm-libs`
	"github.com/generalzgd/grpc-svr-frame/common"
	"github.com/generalzgd/grpc-svr-frame/common/gatepack"
	"github.com/generalzgd/micro-proto/goproto/comm"
	"github.com/generalzgd/micro-proto/goproto/gw"
	u "github.com/generalzgd/micro-proto/goproto/user"
	"github.com/golang/protobuf/proto"

	"github.com/generalzgd/grpc-gw/define"
)

func msgFormatter(bts []byte, cmdId uint16) string {
	if msg, ok := gw.GetMsgObjById(cmdId); ok {
		if err := proto.Unmarshal(bts, msg); err == nil {
			return fmt.Sprintf("[%v]", msg)
		}
	}

	return ""
}

// 给客户端传递消息
func (p *Manager) TransmitProtoDown(info *common.ClientConnInfo, msg proto.Message, cmdId uint16) {
	bts, err := proto.Marshal(msg)
	if err != nil {
		return
	}
	logs.Info("TransmitProtoDown() uid:%d platform:%X Body:%v", info.Uid, info.Platform, msg)
	p.TransmitBytesDown(info, bts, cmdId, nil)
}

func (p *Manager) TransmitBytesDown(info *common.ClientConnInfo, bts []byte, cmdId uint16, formatter func([]byte, uint16) string) {
	if formatter == nil {
		formatter = func(bytes []byte, u uint16) string {
			return ""
		}
	}

	links := p.linkCenter.Get(info.Uid, info.Platform)


	for _, link := range links {

		pack := &gatepack.GateClientPack{
			GateClientPackHead: gatepack.GateClientPackHead{
				Length: uint16(len(bts)),
				CmdId:  cmdId,
			},
			Body: bts,
		}
		p.SendToClient(link, p.AllocPacket(pack))
		// logs.Debug("send to client:", pack.String())
		logs.Debug("TransmitBytesDown uid:%d platform:%x link:%x head:%v Body:%v", info.Uid, info.Platform, libs.GetTarPointer(link), pack.GateClientPackHead, formatter(bts, cmdId))
	}

	if len(links) == 0 {
		logs.Debug("TransmitBytesDown fail， uid: %v check: %v", info.Uid, p.linkCenter.Has(info.Uid))
		logs.Info("TransmitBytesDown fail with no link. uid:%d platform:%X Body:%v", info.Uid, info.Platform, formatter(bts, cmdId))
	} else {
		logs.Debug("TransmitBytesDown uid:%d platform:%x links:%v", info.Uid, info.Platform, len(links))
	}
}

func (p *Manager) OnLogin(ctx context.Context, req *u.LoginNotify) (reply *comm.CommReply, err error) {
	reply = &comm.CommReply{}
	tarInfo, _ := p.GetClientInfo(ctx)
	if tarInfo.Uid != req.User.Uid || tarInfo.Platform != req.User.Platform {
		reply.Code = 1
		return
	}
	// 同一个网关不处理，直接跳过
	if p.gateIp != "" && p.gateIp == tarInfo.GateIp {
		return
	}
	// 在其它网关上，需要踢掉对应端的链接
	elders := p.linkCenter.Get(tarInfo.Uid, tarInfo.Platform)
	for _, elder := range elders {
		if elder.State != nil {
			if elderInfo, ok := elder.State.(*common.ClientConnInfo); ok {
				elderInfo.State = false
			}
		}
		p.SendError(elder, define.NewLoginError)
	}
	return
}

func (p *Manager) KickOut(ctx context.Context, req *u.KickOutRequest) (reply *comm.Error, err error) {
	reply = &comm.Error{}
	tarInfo, _ := p.GetClientInfo(ctx)

	logs.Debug("on Kickout: %v, %v, %v", tarInfo, req.String(), p.gateIp)
	// 同一个网关不处理，直接跳过
	if p.gateIp != "" && p.gateIp == tarInfo.GateIp {
		logs.Debug("on Kickout: same gate ip")
		return
	}
	// 在其它网关上，需要踢掉对应端的链接
	elders := p.linkCenter.Get(tarInfo.Uid, tarInfo.Platform)
	for _, elder := range elders {
		p.SendError(elder, define.NewLoginError)
	}
	return
}

func (p *Manager) Notify(ctx context.Context, req *u.MsgNotify) (*comm.CommReply, error) {
	logs.Debug("Notify() got=[%v]", req.String())

	for _, u := range req.User {
		info := &common.ClientConnInfo{
			Uid:      u.Uid,
			Platform: u.Platform,
		}
		p.TransmitBytesDown(info, req.Message, uint16(req.CmdId), msgFormatter)
	}

	return &comm.CommReply{}, nil
}

// 广播登录事件，踢掉对应的客户端
// func (p *Manager) OnLogin(ctx context.Context, req *ZQProto.ImLoginNotify) (reply *ZQProto.CommReply, err error) {
//	reply = &ZQProto.CommReply{}
//	tarInfo, _ := p.GetClientInfo(ctx)
//	if tarInfo.Uid != req.User.Uid || tarInfo.Platform != req.User.Platform {
//		reply.Code = 1
//		return
//	}
//	// 同一个网关不处理，直接跳过
//	if p.gateIp != "" && p.gateIp == tarInfo.GateIp {
//		return
//	}
//	// 在其它网关上，需要踢掉对应端的链接
//	elders := p.linkCenter.Get(tarInfo.Uid, tarInfo.Platform)
//	for _, elder := range elders {
//		if elder.State != nil {
//			if elderInfo, ok := elder.State.(*common.ClientConnInfo); ok {
//				elderInfo.State = false
//			}
//		}
//		p.SendError(elder, define.NewLoginError)
//	}
//
//	return
// }

// func (p *Manager) Kickout(ctx context.Context, req *ZQProto.ImKickoutRequest) (reply *ZQProto.ImError, err error) {
//	reply = &ZQProto.ImError{}
//	tarInfo, _ := p.GetClientInfo(ctx)
//
//	logs.Debug("on Kickout: %v, %v, %v", tarInfo, req.String(), p.gateIp)
//	// 同一个网关不处理，直接跳过
//	if p.gateIp != "" && p.gateIp == tarInfo.GateIp {
//		logs.Debug("on Kickout: same gate ip")
//		return
//	}
//	// 在其它网关上，需要踢掉对应端的链接
//	elders := p.linkCenter.Get(tarInfo.Uid, tarInfo.Platform)
//	for _, elder := range elders {
//		p.SendError(elder, define.NewLoginError)
//	}
//
//	return
// }
