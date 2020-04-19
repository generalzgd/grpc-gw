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
	`context`
	`crypto/tls`
	`errors`
	`fmt`
	`io`
	`net`
	`net/http`
	`runtime/debug`
	`strings`
	`time`

	`github.com/astaxie/beego/logs`
	libs `github.com/generalzgd/comm-libs`
	`github.com/generalzgd/comm-libs/conf/ymlcfg`
	"github.com/generalzgd/comm-libs/ecode"
	"github.com/generalzgd/comm-libs/env"
	`github.com/generalzgd/comm-libs/mail`
	`github.com/generalzgd/deepcopy/dcopy`
	"github.com/generalzgd/grpc-svr-frame/common"
	"github.com/generalzgd/grpc-svr-frame/common/gatepack"
	lb `github.com/generalzgd/grpc-svr-frame/grpc-consul`
	ctrl "github.com/generalzgd/grpc-svr-frame/grpc-ctrl"
	`github.com/generalzgd/link`
	`github.com/generalzgd/micro-cmdid/gocmd`
	`github.com/generalzgd/micro-proto/goproto/comm`
	`github.com/generalzgd/micro-proto/goproto/gw`
	u `github.com/generalzgd/micro-proto/goproto/user`
	"github.com/golang/protobuf/proto"
	`github.com/grpc-ecosystem/grpc-gateway/runtime`
	`google.golang.org/grpc`
	"google.golang.org/grpc/reflection"

	`github.com/generalzgd/grpc-gw/config`
	"github.com/generalzgd/grpc-gw/define"
	`github.com/generalzgd/grpc-gw/prewarn`
	`github.com/generalzgd/grpc-gw/statistic`
	_ `github.com/generalzgd/grpc-gw/statistic/gorute`
	_ `github.com/generalzgd/grpc-gw/statistic/mem`
	_ `github.com/generalzgd/grpc-gw/statistic/qps`
	_ `github.com/generalzgd/grpc-gw/statistic/tps`
	`github.com/generalzgd/grpc-gw/util`
)

type GateMsgHandler func(session *link.Session, pack gatepack.GateClientPack, info *common.ClientConnInfo) (bool, ecode.IErrCode)

type Manager struct {
	GateClientPackEncrypt
	ctrl.GrpcController
	cfg        config.AppConfig
	wsCfg      config.GateConnConfig
	svrMap     map[string]interface{}
	handlerMap map[uint16]GateMsgHandler
	//
	linkCenter     *define.LinkCenter
	httpInfoCenter *define.ClientInfoCenter
	//
	gateIp string
}

func NewManager() *Manager {
	return &Manager{
		GrpcController: ctrl.MakeGrpcController(),
		svrMap:         map[string]interface{}{},
		handlerMap:     map[uint16]GateMsgHandler{},
		linkCenter:     define.NewLinkCenter(),
		httpInfoCenter: define.NewClientInfoCenter(),
	}
}

func (p *Manager) Init(cfg config.AppConfig) {
	p.cfg = cfg
	pool := util.MakePool(cfg.MemPool.Type,
		cfg.MemPool.MinChunk, cfg.MemPool.MaxChunk, cfg.MemPool.Factor, cfg.MemPool.PageSize)
	p.GateClientPackEncrypt.Init(TcpGatePackageHeadLength, TcpGatePackageMazSize, pool)
	//
	p.handlerMap = map[uint16]GateMsgHandler{
		gocmd.IdHeartbeat:       p.handleHeartBeat,
		gocmd.IdLoginRequest:  p.handleLogin,
		gocmd.IdLogoutRequest: p.handleLogout,
	}

	p.gateIp = libs.GetInnerIp()
	logs.Info(p.cfg.Name, "Ip:", p.gateIp)
	//
	prewarn.SetSendMailCallback(func(body string) {
		// 发送邮件
		title := fmt.Sprintf("%s - %s Warning(%s)", p.cfg.Name, env.GetEnvName(), p.gateIp)
		mail.SendMailByUrl(title, body, cfg.MailAddr...)
	})
	statistic.SetWarnHandler(prewarn.NewWarn)
	lb.InitRegister(p.cfg.Consul.Address)
}

func (p *Manager) Destroy() {
	p.DisposeGrpcConn("")
}

func (p *Manager) ServeClient() error {
	for _, item := range p.cfg.ServeList {
		switch item.Type {
		case define.GwTypeHttp:
			if err := p.serveHttp(item); err != nil {
				return err
			}
		case define.GwTypeTcp:
			if err := p.serveTcp(item); err != nil {
				return err
			}
		case define.GwTypeUdp:
		case define.GwTypeWs:
			if err := p.serveWs(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// 启动下行通讯端口
func (p *Manager) ServeGrpc() error {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(1000),
		grpc.MaxRecvMsgSize(32 * 1024),
		grpc.MaxSendMsgSize(32 * 1024),
		grpc.ReadBufferSize(8 * 1024),
		grpc.WriteBufferSize(8 * 1024),
		grpc.ConnectionTimeout(5 * time.Second),
	}
	// todo tls

	addr := fmt.Sprintf(":%d", p.cfg.RpcSvr.Port) // p.cfg.RpcSvr.Address[0]
	s := grpc.NewServer(opts...)
	gw.RegisterGateWayDownServer(s, p)
	// ZQProto.RegisterHfGateWayDownServer(s, p)
	reflection.Register(s)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logs.Error("failed to listen: %v", err)
	}
	go func() {
		if err := s.Serve(lis); err != nil {
			logs.Error("failed to serve: %v", err)
		}
	}()
	logs.Debug("start serve grpc.", addr, p.cfg.RpcSvr.Secure)
	return nil
}

//
func (p *Manager) serveHttp(cfg config.GateConnConfig) error {
	addr := fmt.Sprintf(":%d", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	// todo 目标服务是否启用tls

	// 测试代码
	// p.httpInfoCenter.Put("12F44847-85EB-E6D8-6D38-09AA118BB528", &common.ClientConnInfo{
	// 	Uid:      163,
	// 	Platform: 1,
	// 	State:    true,
	// })

	err := gw.RegisterGatewayHandlerClient(ctx, mux, opts, p.getEndpointAddrByMeth, p.getGrpcClientConnByMeth,
		p.httpCallBeginHandler, p.httpCallDoneHandler, p.httpCallQps)
	if err != nil {
		logs.Error("serve http gate fail.", err)
		return err
	}

	go func() {
		http.ListenAndServe(addr, mux)
	}()
	// p.svrMap[cfg.Name] = server
	logs.Info("start serve %s(%s) with secure(%v)", cfg.Name, addr, cfg.Secure)
	return nil
}

// grpc转发前的回调处理
func (p *Manager) httpCallBeginHandler(meth string, req *http.Request) (int, bool) {
	statistic.NewRecord(statistic.StatTps)
	// 校验cookie
	cookieGuid, err := req.Cookie("ZQ_GUID")
	if err != nil || cookieGuid.Value == "" /*|| strings.Index(cookie_guid.Value, ".") < 0*/ {
		return http.StatusUnauthorized, false
	}
	// 映射到对应cmdid, meth->package.Service/Method，例如：ZQProto.Authorize/Login
	cmdId := gw.GetIdByMeth(meth) // Method2Cmdid[meth]
	if cmdId == gocmd.IdLoginRequest {
		// 不处理 conninfo
	} else {
		info, ok := p.httpInfoCenter.Get(cookieGuid.Value)
		// 如果未登录，过期，则终止转发
		if !ok || !info.State {
			return http.StatusUnauthorized, false
		}

		if cmdId == gocmd.IdLogoutRequest {
			p.httpInfoCenter.Delete(cookieGuid.Value, nil)
		}
		// 将客户端链接信息转换为grpc的Header信息
		tmp, _ := dcopy.InstanceToMap(info)
		for k, v := range tmp {
			// header key需要有MetadataHeaderPrefix开头，才会识别出来并转发给endpoint; 终端取到的key不含MetadataHeaderPrefix
			req.Header.Add(runtime.MetadataHeaderPrefix+k, libs.Interface2String(v)) // MetadataHeaderPrefix
		}
	}
	return http.StatusAccepted, true
}

// grpc结束回调处理
func (p *Manager) httpCallDoneHandler(meth string, reply proto.Message, w http.ResponseWriter, req *http.Request) {
	// 校验cookie
	cookieGuid, err := req.Cookie("ZQ_GUID")
	if err != nil || cookieGuid.Value == "" {
		return
	}
	// 映射到对应cmdid, meth->package.Service/Method，例如：ZQProto.Authorize/Login
	cmdid := gw.GetIdByMeth(meth) // Method2Cmdid[meth]
	if cmdid == gocmd.IdLoginRequest {
		// 登录成功，保存用户信息
		if data, ok := reply.(*u.LoginReply); ok {
			info, ok := p.httpInfoCenter.Get(cookieGuid.Value)
			if !ok {
				info = &common.ClientConnInfo{}
			}
			if data.Code == 0 {
				p.saveUserInfo(info, data, cookieGuid.Value, 3600*24) // 1 天后过期
			} else {
				info.State = false
			}
			p.httpInfoCenter.Put(cookieGuid.Value, info)
		}
	}
}

func (p *Manager) httpCallQps(d time.Duration) {
	statistic.NewRecord(statistic.StatQps, d)
}

func (p *Manager) serveTcp(cfg config.GateConnConfig) error {
	addr := fmt.Sprintf(":%d", cfg.Port)
	var server *link.Server
	if cfg.Secure {
		tlsCfg, err := p.GetTlsConfig(cfg.CertFiles...)
		if err != nil {
			return err
		}
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		lis := tls.NewListener(ln, tlsCfg)
		server = link.NewServer(
			lis,
			link.ProtocolFunc(func(rw io.ReadWriter) (link.Codec, error) {
				return newTlsCodec(rw.(net.Conn), cfg.BufferSize, &p.GateClientPackEncrypt)
			}),
			cfg.SendChanSize,
			link.HandlerFunc(func(session *link.Session) {
				p.handleSession(session, cfg.MaxConn, cfg.IdleTimeout, `tls`)
			}),
		)
	} else {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		server = link.NewServer(
			lis,
			link.ProtocolFunc(func(rw io.ReadWriter) (link.Codec, error) {
				return newTcpCodec(rw.(net.Conn), cfg.BufferSize, &p.GateClientPackEncrypt)
			}),
			cfg.SendChanSize,
			link.HandlerFunc(func(session *link.Session) {
				p.handleSession(session, cfg.MaxConn, cfg.IdleTimeout, `tcp`)
			}),
		)
	}

	go func() {
		if err := server.Serve(); err != nil {
			logs.Error("tcp gate way serve error.", err)
		}
	}()
	p.svrMap[cfg.Name] = server
	logs.Info("start serve %s(%s) with secure(%v)", cfg.Name, addr, cfg.Secure)
	return nil
}

func (p *Manager) serveWs(cfg config.GateConnConfig) error {
	var (
		httpServeMux = http.NewServeMux()
		err          error
	)
	httpServeMux.HandleFunc("/", p.serveWebSocket)
	httpServeMux.HandleFunc("/health", p.serveWebHealth)
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{Addr: addr, Handler: httpServeMux}
	server.SetKeepAlivesEnabled(true)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if cfg.Secure {
		tlsCfg, err := p.GetTlsConfig(cfg.CertFiles...)
		if err != nil {
			return err
		}
		ln = tls.NewListener(ln, tlsCfg)
	}

	go func() {
		if err = server.Serve(ln); err != nil {
			logs.Info("start serve %s(%s) with secure(%s)", cfg.Name, addr, cfg.Secure)
		}
	}()

	p.wsCfg = cfg
	p.svrMap[cfg.Name] = server
	logs.Info("start serve %s(%s) with secure(%v)", cfg.Name, addr, cfg.Secure)
	return nil
}

/*
* todo 阿里云slb的健康检测
 */
func (p *Manager) serveWebHealth(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(201)
}

/*
* todo websocket处理
 */
func (p *Manager) serveWebSocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		logs.Error("Websocket Upgrade error(%v), userAgent(%v), reqHeader(%v)", err, req.UserAgent(), req.Header)
		return
	}

	var (
		lAddr = ws.LocalAddr()
		rAddr = ws.RemoteAddr()
	)
	// nginx代理的客户端ip获取
	realIp := req.Header.Get("X-Real-Ip")

	if len(realIp) == 0 {
		// slb的客户端ip获取。X-Forwarded-For: 用户真实IP, 代理服务器1-IP， 代理服务器2-IP，...
		tmp := req.Header.Get("X-Forwarded-For")
		splitIdx := strings.Index(tmp, ",")
		if splitIdx > 0 {
			realIp = tmp[:splitIdx]
		} else {
			realIp = tmp
		}
	}

	session := link.NewSession(newWssCodec(ws, &p.GateClientPackEncrypt), p.wsCfg.SendChanSize)

	session.SetRealIp(realIp)

	logs.Debug("websocket new connection", lAddr, "with", rAddr, ">>", realIp)

	p.handleSession(session, p.wsCfg.MaxConn, p.wsCfg.IdleTimeout, `ws`)
}

func (p *Manager) handleSession(session *link.Session, maxConn int, idleTimeout time.Duration, netType string) {
	conn := session.Codec()
	clientIp := session.GetRealIp()
	id := conn.SocketID()

	// 判断连接是否来自健康检测,如果是健康检测的,就不打印错误日志了
	isHealthCheckClient := false
	for _, hip := range p.cfg.HealthCheckIp {
		if clientIp[0:len(hip)] == hip {
			isHealthCheckClient = true
			break
		}
	}

	info := &common.ClientConnInfo{
		SocketId: id,
		ClientIp: clientIp,
		GateIp:   p.gateIp,
	}
	session.State = info
	logs.Info("session connected ip:%s sid:%d isHealthCheck:%v netType:%s", clientIp, id, isHealthCheckClient, netType)
	// 底层错误信息句柄
	var zrr ecode.IErrCode
	defer func() {
		p.onUserLogout(info, zrr, session)
		if info.State {
			info.State = false
			except := false
			if zrr == nil {
				except = true
			}
			go p.justCallAuthSvrLogout(session, info, except)
			logs.Info("session user logout:%v", info)
		}
		logs.Info("session disconnected ip:%s sid:%d netType:%s", clientIp, id, netType)
		if r := recover(); r != nil {
			logs.Error("session panic ip:%s sid:%d, err:%v, stack:%s", clientIp, id, r, string(debug.Stack()))
		}
		session.State = nil
	}()

	for {
		if idleTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(idleTimeout)); err != nil {
				zrr = define.LinkError.SetError(err)
				if !isHealthCheckClient {
					logs.Error("session timeout error. ip:%s sid:%d err:%v netType:%s", clientIp, id, err, netType)
				}
				return
			}
		}
		buf, err := session.Receive()
		if err != nil {
			zrr = define.LinkError.SetError(err)
			if !isHealthCheckClient {
				logs.Error("session receive error:%v. ip:%s sid:%d err:%v netType:%s", err, clientIp, id, err, netType)
			}
			return
		}

		data := *(buf.(*[]byte))
		packet := gatepack.GateClientPack{}
		zrr := p.DecodePacket(data, &packet)
		p.Free(data)
		if zrr != nil {
			if !isHealthCheckClient {
				logs.Error("session decode packet error. ip:%s sid:%d err:%v netType:%s", clientIp, id, zrr, netType)
			}
			return
		}

		// if packet.Cmdid != gocmd.ID_Heartbeat {
		// 	logs.Debug("session receive packet ip:%s sid:%d head:%v body:%v", clientIp, id, packet.GateClientPackHead, p.formatProtoStr(packet.Cmdid, packet.Body) )
		// }

		// logs.Debug("receive pack:", fmt.Sprintf("head:%v,Body:%v", packet.GateClientPackHead, p.formatProtoStr(packet.Cmdid, packet.Body)))
		// 统计tps
		statistic.NewRecord(statistic.StatTps)
		// 保存时间戳
		// info.Timestamp = time.Now().UnixNano()
		// 生成traceId
		// info.TraceId = fmt.Sprintf(`%s-%d`, info.Guid, info.Timestamp)
		// 预处理登录，登出，心跳，状态判断; 由于里面已经做了错误处理，所以这里采用临时变量zr,
		if goon, zr := p.preprocessPack(packet, session, info); zr != nil {
			logs.Error("session process pack error. ip:%s sid:%d err:%v", clientIp, id, zr)
			return
		} else {
			if !goon {
				continue
			}
		}
		// 	转发协议，转发失败则发送错误
		go func(session *link.Session, pack gatepack.GateClientPack, info *common.ClientConnInfo) {
			if err := p.transmitPack(session, packet, info); err != nil {
				logs.Error("session transmit pack error. ip:%s sid:%d err:%v", clientIp, id, err)
			}
		}(session, packet, info)
	}
}

func (p *Manager) formatProtoStr(cmdId uint16, body []byte) string {
	if obj, ok := gw.GetMsgObjById(cmdId); ok {
		if err := proto.Unmarshal(body, obj); err == nil {
			return libs.GetStructName(obj) + "{" + obj.String() + "}"
		}
	}
	return ""
}

func (p *Manager) getEndpointByCmdId(cmdId uint16) (ymlcfg.EndpointConfig, bool) {
	meth := gw.GetMethById(cmdId)
	return p.getEndpointByMeth(meth)
}

func (p *Manager) getEndpointByMeth(meth string) (ymlcfg.EndpointConfig, bool) {
	_, tarSvr, _, _ := gw.ParseMethod(meth)
	tarSvr = strings.ToLower(tarSvr + "svr")
	cfg, ok := p.cfg.EndpointSvr[tarSvr]
	return cfg, ok
}

func (p *Manager) getEndpointAddrByMeth(meth string) string {
	_, tarSvr, _, _ := gw.ParseMethod(meth)
	tarSvr = strings.ToLower(tarSvr + "svr")
	cfg, ok := p.cfg.EndpointSvr[tarSvr]
	if ok {
		return cfg.Address
	}
	return ""
}

func (p *Manager) getGrpcClientConnByMeth(meth string) (*grpc.ClientConn, func(), error) {
	_, tarSvr, _, _ := gw.ParseMethod(meth)
	tarSvr = strings.ToLower(tarSvr + "svr")
	if cfg, ok := p.cfg.EndpointSvr[tarSvr]; ok {
		// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		return p.GetGrpcConnWithLB(cfg)
	}
	return nil, nil, errors.New("endpoint config empty:" + tarSvr)
}

func (p *Manager) SendError(session *link.Session, zrr ecode.IErrCode) {
	msg := &comm.Error{
		Code:    comm.ErrorType(zrr.GetCode()),
		Message: zrr.GetErrMsg(),
	}
	bts, err := proto.Marshal(msg)
	if err != nil {
		logs.Error("send kickout msg err.", err)
		return
	}
	pack := &gatepack.GateClientPack{
		GateClientPackHead: gatepack.GateClientPackHead{
			Length: uint16(len(bts)),
			CmdId:  gocmd.IdError,
		},
		Body: bts,
	}
	if err := p.SendToClient(session, p.AllocPacket(pack)); err != nil {
		logs.Error("send error msg err.", err)
		return
	}
	logs.Debug("session send error packet sid:%v {%v}:{%v}", session.Codec().SocketID(), pack.GateClientPackHead, msg.String())
	time.AfterFunc(time.Second, func() {
		if session != nil {
			if err := session.Close(); err != nil {
				logs.Error("close client session err.", err)
			}
			logs.Debug("close link sid:", session.Codec().SocketID())
		}
	})
}

// 统一封装session 发送方法
func (p *Manager) SendToClient(session *link.Session, msg []byte) error {
	err := session.Send(&msg)
	if err != nil {
		p.Free(msg)
		session.Close()
	}
	// 下行也要记录tps
	statistic.NewRecord(statistic.StatTps)
	return err
}
