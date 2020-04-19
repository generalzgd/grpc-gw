/**
 * @version: 1.0.0
 * @author: zhangguodong:general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: def.go
 * @time: 2019/8/5 13:48
 */
package define

import (
	`errors`

	`github.com/generalzgd/comm-libs/ecode`
	`github.com/generalzgd/micro-proto/goproto/comm`
	`github.com/toolkits/slice`
)

const (
	GwTypeHttp = "http"
	GwTypeWs   = "ws"
	GwTypeTcp  = "tcp"
	GwTypeUdp  = "udp"
)

const (
// NewLoginStr  = "重复登录"
// LinkBreakStr = "链接断开"
)

var (
	NilErr              = errors.New("target is nil")
	LinkError           = ecode.MakeErr("link error", "链接错误", int(comm.ErrorType_LinkErr))
	UnknowError         = ecode.MakeErr("unknow error", "未知错误", int(comm.ErrorType_UnknowErr))
	NilError            = ecode.MakeFromErr(NilErr).SetErrMsg("空指针").SetCode(int(comm.ErrorType_NilErr))
	GwTypeErr           = errors.New("gate type error")
	TooLargeErr         = errors.New("too large packet")
	TooLargeError       = ecode.MakeFromErr(TooLargeErr).SetErrMsg("协议包过大").SetCode(int(comm.ErrorType_PackTooLarge))
	PackVerErr          = errors.New("packet version error")
	PackVerError        = ecode.MakeFromErr(PackVerErr).SetErrMsg("协议包版本错误").SetCode(int(comm.ErrorType_PackVerFail))
	CmdidFieldErr       = errors.New("cmdid file invalidate")
	CmdidFieldError     = ecode.MakeFromErr(CmdidFieldErr).SetErrMsg("Cmdid字段错误").SetCode(int(comm.ErrorType_CmdidFail))
	SerializeError      = ecode.MakeErr("", "协议序列化错误", int(comm.ErrorType_SerializeFail))
	UnserializeError    = ecode.MakeErr("", "协议反序列化错误", int(comm.ErrorType_UnserializeFail))
	EncryptError        = ecode.MakeErr("", "加密错误", int(comm.ErrorType_EncryptFail))
	DecryptError        = ecode.MakeErr("", "解密错误", int(comm.ErrorType_DecryptFail))
	CompressError       = ecode.MakeErr("", "压缩错误", int(comm.ErrorType_CompressFail))
	UncompressError     = ecode.MakeErr("", "解压错误", int(comm.ErrorType_UncompressFail))
	MarshalError        = ecode.MakeErr("codec set error", "编码错误", int(comm.ErrorType_MarshalFail))
	EndpointError       = ecode.MakeErr("endpoint error", "终端异常", int(comm.ErrorType_EndpointFail))
	ClientMetadataError = ecode.MakeErr("", "客户端元信息错误", int(comm.ErrorType_ClientInfoFail))
	TransmitError       = ecode.MakeErr("", "网关传输错误", int(comm.ErrorType_TransmitFail))
	NotLoginError       = ecode.MakeErr("not login yet", "请先登录", int(comm.ErrorType_NotLogin))
	NewLoginError       = ecode.MakeErr("new login", "重复登录", int(comm.ErrorType_NewLogin))
	HasLoginError       = ecode.MakeErr("has login", "已经登录", int(comm.ErrorType_HasLogin))
)

var (
	GwTypeList = []string{GwTypeHttp, GwTypeWs, GwTypeTcp, GwTypeUdp}
)

func ValidateGwType(in string) bool {
	return slice.ContainsString(GwTypeList, in)
}
