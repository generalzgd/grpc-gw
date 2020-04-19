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
	"context"
	`errors`
	`fmt`
	`runtime/debug`
	`strings`
	`time`

	`github.com/astaxie/beego/logs`
	libs `github.com/generalzgd/comm-libs`
	"github.com/generalzgd/comm-libs/ecode"
	"github.com/generalzgd/grpc-svr-frame/common"
	"github.com/generalzgd/grpc-svr-frame/common/gatepack"
	`github.com/generalzgd/link`
	"github.com/generalzgd/micro-cmdid/gocmd"
	`github.com/generalzgd/micro-proto/goproto/comm`
	`github.com/generalzgd/micro-proto/goproto/gw`
	u `github.com/generalzgd/micro-proto/goproto/user`
	`github.com/golang/protobuf/proto`
	`google.golang.org/grpc/metadata`

	"github.com/generalzgd/grpc-gw/define"
	`github.com/generalzgd/grpc-gw/statistic`
)

func (p *Manager) onUserLogin(info *common.ClientConnInfo, session *link.Session) {
	// 保存客户端信息
	p.httpInfoCenter.Put(info.Guid, info)
	// 保存链接信息
	elder := p.linkCenter.Put(info.Uid, info.Platform, session)
	if elder != nil {
		logs.Error("onUserLogin() close olderLink=[%x] newLink=[%x]", libs.GetTarPointer(elder),
			libs.GetTarPointer(session))
		if elder.State != nil {
			if elderInfo, ok := elder.State.(*common.ClientConnInfo); ok {
				elderInfo.State = false
			}
		}
		// 关闭老的客户端连接
		p.SendError(elder, define.NewLoginError)
		// logs.Info("elder sid:%v, current sid:%v for %v %v", elder.Codec().SocketID(), session.Codec().SocketID(), info.Uid, info.Platform)
	} else {
		logs.Info("onUserLogin() no elder link uid=[%d] platform=[%x] link=[%x]", info.Uid, info.Platform, libs.GetTarPointer(session))
	}

	// todo 发给brosvr, 然后再发给gateway
	cfg := p.cfg.EndpointSvr[BroSvr]
	// address := "consul:///brosvr:8080/OnLogin" // cfg.Address[0] // .ImNotifySvr.Address[0]

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, fun, err := p.GetGrpcConnWithLB(cfg)
	if err != nil {
		logs.Error("onUserLogin() got err=[%v]", err)
		return
	}
	defer fun() // 归还给连接池

	client := gw.NewBroClient(conn)

	req := &u.LoginNotify{
		User: &u.OneUser{
			Uid:      info.Uid,
			Platform: info.Platform,
		},
		LoginTime:     info.LoginTime,
		SocketId:      info.SocketId,
		CurrentGateIp: info.GateIp,
	}

	ctx = p.MakeOutgoingContextByClientInfo(context.Background(), info)
	rep, err := client.OnLogin(ctx, req)
	if err != nil {
		logs.Error("onUserLogin() err=[%v] uid=[%v]", err, info.Uid)
	} else {
		logs.Debug("onUserLogin() ok req=[%v] uid=[%v]", rep, info.Uid)
	}
}

// 断开客户端链接
func (p *Manager) onUserLogout(info *common.ClientConnInfo, zrr ecode.IErrCode, currentSession *link.Session) {
	// 删除客户端信息
	p.httpInfoCenter.Delete(info.Guid, info)

	elders := p.linkCenter.Delete(info.Uid, info.Platform, currentSession)
	if zrr != nil {
		for _, elder := range elders {
			p.SendError(elder, zrr)
		}
	}

	if zrr == nil || len(elders) == 0 {
		currentSession.Close()
	}
}

// 发送响应包，包头沿用了旧包头，即复用了seq，opt字段
func (p *Manager) sendReplyPack(session *link.Session, pack gatepack.GateClientPack, reply proto.Message) {
	var bts []byte
	var err error

	bts, err = gw.EncodeBytes(pack.Codec, reply)
	if err != nil {
		return
	}

	pack.CmdId = gw.GetIdByMsgObj(reply)
	pack.Length = uint16(len(bts))
	pack.Body = bts
	p.SendToClient(session, p.AllocPacket(&pack))

	conn := session.Codec()
	clientIp := session.GetRealIp()
	id := conn.SocketID()
	logs.Debug("session send packet ip:%s sid:%d %v:%v>>{%v}", clientIp, id, pack.GateClientPackHead, libs.GetStructName(reply), reply.String())
}

// 预处理 登录，登出，心跳，状态判断
func (p *Manager) preprocessPack(pack gatepack.GateClientPack, session *link.Session, info *common.ClientConnInfo) (goon bool, err ecode.IErrCode) {
	begin := time.Now()
	defer func() {
		statistic.NewRecord(statistic.StatQps, time.Since(begin))
		if r := recover(); r != nil {
			logs.Error("session process pack panic.", r, string(debug.Stack()))
			switch x := r.(type) {
			case error:
				err = define.NilError.SetError(x)
			case string:
				err = define.NilError.SetError(errors.New(x))
			default:
				err = define.UnknowError
			}
		}
		// 异常返回
		if err != nil {
			ec := &comm.Error{
				Code:    comm.ErrorType(err.GetCode()),
				Message: fmt.Sprintf("%s(%v)", err.GetErrMsg(), err.GetError()),
			}
			p.sendReplyPack(session, pack, ec)
		}
	}()

	// 未登陆前，只处理login, heartbeat两个消息
	if pack.CmdId > 0 {
		if handler, ok := p.handlerMap[pack.CmdId]; ok {
			goon, err = handler(session, pack, info)
			return
		}
	} else {
		return false, define.CmdidFieldError
	}
	// 未登陆的情况下， 发送其他协议的话，则拒绝发送
	if info.State == false {
		return false, define.NotLoginError
	}
	return true, nil
}

func (p *Manager) handleHeartBeat(session *link.Session, pack gatepack.GateClientPack, info *common.ClientConnInfo) (bool, ecode.IErrCode) {
	// 发回给客户端
	p.SendToClient(session, p.AllocPacket(&pack))
	return false, nil
}

func (p *Manager) handleLogin(session *link.Session, pack gatepack.GateClientPack, info *common.ClientConnInfo) (bool, ecode.IErrCode) {
	if info.State {
		logs.Debug("user has login")
		return false, define.HasLoginError
	}

	method := gw.GetMethById(pack.CmdId)
	cfg, ok := p.getEndpointByMeth(method)
	if !ok {
		return false, define.EndpointError
	}

	loginReq := &u.LoginRequest{}
	if err := gw.DecodeBytes(pack.Body, pack.Codec, loginReq); err != nil {
		return false, define.MarshalError.SetError(err)
	}

	info.Platform = loginReq.Platform
	info.Uid = loginReq.Uid

	md, err := p.ClientInfoToMD(info)
	if err != nil {
		return false, define.ClientMetadataError
	}

	doneHandler := func(reply proto.Message) {
		if data, ok := reply.(*u.LoginReply); ok && data.Code == 0 {
			p.saveUserInfo(info, data, loginReq.Token, 0)
			//
			logs.Debug("handleLogin() got=[%v]", data)
			if data.LoginCnt > 1 {
				kickOut := &u.KickOutRequest{
					Uid:          info.Uid,
					LastSid:      info.SocketId,
					LastPlatform: info.Platform,
				}
				go p.kickOutOldLink(info, kickOut)
			}
			go p.onUserLogin(info, session)
		} else {
			info.State = false
		}
		p.sendReplyPack(session, pack, reply)
	}

	// dialOption := p.GetDialOption(cfg) // grpc.WithInsecure()
	logs.Debug("start call grpc:", cfg.Address, fmt.Sprintf("head:%v,Body:%v", pack.GateClientPackHead, p.formatProtoStr(pack.CmdId, pack.Body)))

	// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, fun, err := p.GetGrpcConnWithLB(cfg)
	if err != nil {
		return false, define.TransmitError.SetError(err)
	}
	defer fun() // 归还

	args := &gw.TransmitArgs{
		Method:       method,
		Endpoint:     cfg.Address,
		Conn:         conn,
		MD:           md,
		Data:         pack.Body,
		Codec:        pack.Codec,
		DoneCallback: doneHandler,
		Opts:         nil,
	}

	if err := gw.RegisterTransmitor(args); err != nil {
		logs.Error("transmit fail.", method, loginReq.String())
		return true, define.TransmitError.SetError(err)
	}
	return false, nil
}

func (p *Manager) handleLogout(session *link.Session, pack gatepack.GateClientPack, info *common.ClientConnInfo) (bool, ecode.IErrCode) {
	method := gw.GetMethById(pack.CmdId)
	cfg, ok := p.getEndpointByMeth(method)
	if !ok {
		return false, define.EndpointError
	}

	md, err := p.ClientInfoToMD(info)
	if err != nil {
		return false, define.ClientMetadataError
	}

	doneHandler := func(reply proto.Message) {
		info.State = false
		p.sendReplyPack(session, pack, reply)
		time.AfterFunc(time.Second, func() {
			session.Close()
		})
	}

	// dialOption := p.GetDialOption(cfg) // grpc.WithInsecure()
	logs.Debug("start call grpc:", cfg.Address, fmt.Sprintf("head:%v,Body:%v", pack.GateClientPackHead, p.formatProtoStr(pack.CmdId, pack.Body)))

	// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, fun, err := p.GetGrpcConnWithLB(cfg)
	if err != nil {
		return false, define.TransmitError.SetError(err)
	}
	defer fun()

	args := &gw.TransmitArgs{
		Method:       method,
		Endpoint:     cfg.Address,
		Conn:         conn,
		MD:           md,
		Data:         pack.Body,
		Codec:        pack.Codec,
		DoneCallback: doneHandler,
		Opts:         nil,
	}

	if err := gw.RegisterTransmitor(args); err != nil {
		return true, define.TransmitError.SetError(err)
	}
	return false, nil
}

// 转换协议并发送, 前提是解析出当前的包
func (p *Manager) transmitPack(session *link.Session, pack gatepack.GateClientPack, info *common.ClientConnInfo) (err ecode.IErrCode) {
	begin := time.Now()
	defer func() {
		statistic.NewRecord(statistic.StatQps, time.Since(begin))
		if r := recover(); r != nil {
			logs.Error("session process pack panic.", r, string(debug.Stack()))
			switch x := r.(type) {
			case error:
				err = define.NilError.SetError(x)
			case string:
				err = define.NilError.SetError(errors.New(x))
			default:
				err = define.UnknowError
			}
		}
		// 异常返回
		if err != nil {
			hferr := &comm.Error{
				Code:    comm.ErrorType(err.GetCode()),
				Message: fmt.Sprintf("%s(%v)", err.GetErrMsg(), err.GetError()),
			}
			p.sendReplyPack(session, pack, hferr)
		}
	}()
	var md metadata.MD
	// 根据cmdid映射，得到对应的后端方法名称 package.Service/Method, 例如：ZQProto.Authorize/Login
	meth := gw.GetMethById(pack.CmdId)
	if len(meth) < 1 {
		err = define.CmdidFieldError
		return
	}

	cfg, ok := p.getEndpointByMeth(meth)
	if !ok {
		err = define.EndpointError
		return
	}

	// 转换用户链接信息为metadata.MD，用于grpc的header传输。
	// 接收方将header信息转换成ClientConnInfo结构体，以获得用户链接信息
	if m, e := p.ClientInfoToMD(info); e == nil {
		md = m
	} else {
		err = define.ClientMetadataError.SetError(e)
		return
	}
	// grpc传输结束时的方法调用
	doneHandler := func(reply proto.Message) {
		p.sendReplyPack(session, pack, reply)
	}
	logs.Debug("start call grpc:", cfg.Address, fmt.Sprintf("head:%v,Body:%v", pack.GateClientPackHead, p.formatProtoStr(pack.CmdId, pack.Body)))

	// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, fun, e := p.GetGrpcConnWithLB(cfg)
	if e != nil {
		err = define.TransmitError.SetError(e)
		return
	}
	defer fun()

	args := &gw.TransmitArgs{
		Method:       meth,
		Endpoint:     cfg.Address,
		Conn:         conn,
		MD:           md,
		Data:         pack.Body,
		Codec:        pack.Codec,
		DoneCallback: doneHandler,
		Opts:         nil,
	}
	// 将pack的信息，转换传输给后端的服务
	if e := gw.RegisterTransmitor(args); e != nil {
		err = define.TransmitError.SetError(e)
		return
	}
	return
}

func (p *Manager) justCallAuthSvrLogout(session *link.Session, info *common.ClientConnInfo, except bool) bool {
	method := gw.GetMethById(gocmd.IdLogoutRequest)
	cfg, ok := p.getEndpointByMeth(method)
	if !ok {
		return false
	}

	md, err := p.ClientInfoToMD(info)
	if err != nil {
		return false
	}

	doneHandler := func(reply proto.Message) {
		info.State = false
	}
	// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, fun, err := p.GetGrpcConnWithLB(cfg)
	if err != nil {
		return false
	}
	defer fun()

	req := &u.LogoutRequest{
		TraceId:              "",
		Except:               except,
	}
	bts, _ := proto.Marshal(req)

	args := &gw.TransmitArgs{
		Method:       method,
		Endpoint:     cfg.Address,
		Conn:         conn,
		MD:           md,
		Data:         bts,
		Codec:        gatepack.PackCodecProto,
		DoneCallback: doneHandler,
		Opts:         nil,
	}

	logs.Debug("start call grpc:", cfg.Address)
	if err := gw.RegisterTransmitor(args); err != nil {
		return true
	}
	return false
}

func (p *Manager) kickOutOldLink(info *common.ClientConnInfo, req *u.KickOutRequest) {
	// todo 发给HfBroSvr, 然后再发给gate
	cfg := p.cfg.EndpointSvr[BroSvr]
	// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	conn, fun, err := p.GetGrpcConnWithLB(cfg)
	if err != nil {
		return
	}
	defer fun()

	client := gw.NewBroClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = p.MakeOutgoingContextByClientInfo(ctx, info)
	reply, err := client.KickOut(ctx, req)
	if err != nil {
		return
	}
	logs.Debug("kickOut other link ok.", reply.String())
}

func (p *Manager) saveUserInfo(info *common.ClientConnInfo, from *u.LoginReply, token string, exp int64) {
	info.State = true
	if from.User != nil {
		info.Uid = from.User.Uid
		info.Nickname = from.User.Nickname
	}
	info.Guid = token
	if strings.Index(token, ".") > 0 {
		info.Guid = strings.Split(token, ".")[1]
	}
	info.LoginTime = time.Now().Unix()
	info.Expire = exp // 0:永不过期
}
