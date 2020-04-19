/**
 * @version: 1.0.0
 * @author: zhangguodong:general_zgd
 * @license: LGPL v3
 * @contact: general_zgd@163.com
 * @site: github.com/generalzgd
 * @software: GoLand
 * @file: codec.go
 * @time: 2019/8/6 14:50
 */
package mgr

import (
	`bufio`
	`encoding/binary`
	`io`
	`net`
	`net/http`
	`sync/atomic`
	`time`

	`github.com/funny/slab`
	"github.com/generalzgd/comm-libs/ecode"
	"github.com/generalzgd/grpc-svr-frame/common/gatepack"
	`github.com/generalzgd/link`
	`github.com/gorilla/websocket`

	`github.com/generalzgd/grpc-gw/define`
)

const (
	TcpGatePackageHeadLength = 12
	TcpGatePackageMazSize    = 1024 * 32
)

var (
	connSeed = uint32(0)
	//
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func newTlsCodec(conn net.Conn, bufferSize int, encryptor *GateClientPackEncrypt) (link.Codec, error) {
	id := atomic.AddUint32(&connSeed, 1)
	c := &TlsGateCodec{
		GateCodecBase: GateCodecBase{
			GateClientPackEncrypt: encryptor,
			id:                    id,
			conn:                  conn,
			reader:                bufio.NewReaderSize(conn, bufferSize),
		},
	}
	c.headBuf = c.headDat[:]
	return c, nil
}

func newTcpCodec(conn net.Conn, bufferSize int, encryptor *GateClientPackEncrypt) (link.Codec, error) {
	id := atomic.AddUint32(&connSeed, 1)
	c := &TcpGateCodec{
		GateCodecBase: GateCodecBase{
			GateClientPackEncrypt: encryptor,
			id:                    id,
			conn:                  conn,
			reader:                bufio.NewReaderSize(conn, bufferSize),
		},
	}
	c.headBuf = c.headDat[:]
	return c, nil
}

func newWssCodec(conn *websocket.Conn, encryptor *GateClientPackEncrypt) *WsGateCodec {
	id := atomic.AddUint32(&connSeed, 1)
	c := &WsGateCodec{
		GateClientPackEncrypt: encryptor,
		id:                    id,
		conn:                  conn,
	}
	c.headBuf = c.headDat[:]
	return c
}

// ********************************
type PackEncrypt struct {
	HeadSize      int
	MaxPacketSize int
	Pool          slab.Pool
}

func (p *PackEncrypt) Init(headSize, maxSize int, pool slab.Pool) {
	p.HeadSize = headSize
	p.MaxPacketSize = maxSize
	p.Pool = pool
}

func (p *PackEncrypt) MemSet(mem []byte, v byte) {
	l := len(mem)
	for i := 0; i < l; i++ {
		mem[i] = v
	}
}

func (p *PackEncrypt) Alloc(size int) []byte {
	if p.Pool != nil {
		return p.Pool.Alloc(size)
	}
	return make([]byte, size)
}

func (p *PackEncrypt) Free(msg []byte) {
	if p.Pool != nil {
		p.Pool.Free(msg)
	}
}

func (p *PackEncrypt) AllocPacket(pack interface{}) []byte {
	return []byte{}
}
func (p *PackEncrypt) DecodePacket(data []byte, tar *gatepack.GateClientPack) zqerr.IErrCode {
	return nil
}

// 客户端包
type GateClientPackEncrypt struct {
	PackEncrypt
}

func (p *GateClientPackEncrypt) handleEncrypt(pack *gatepack.GateClientPack) {
	// pack.Encrypt = 0
	// 	todo something
}

func (p *GateClientPackEncrypt) handleDecrypt(pack *gatepack.GateClientPack) {

}

func (p *GateClientPackEncrypt) handleCompress(pack *gatepack.GateClientPack) {
	// todo something
}

func (p *GateClientPackEncrypt) handleUncompress(pack *gatepack.GateClientPack) error {
	// todo something
	return nil
}

func (p *GateClientPackEncrypt) AllocPacket(pkt interface{}) []byte {
	pack, ok := pkt.(*gatepack.GateClientPack)
	if !ok {
		return []byte{}
	}

	pack.Ver = gatepack.ProtocolVer
	pack.Codec = gatepack.PackCodecProto
	// pack.Opt = 0
	// 先加密
	p.handleEncrypt(pack)
	// 压缩
	p.handleCompress(pack)

	buffer := p.Alloc(p.HeadSize + int(pack.Length)) // 重新设置len为0

	_, err := pack.SerializeWithBuf(buffer[:0])
	if err != nil {
		p.Free(buffer)
		return nil
	}
	return buffer
}

// decodePacket decodes gateway message
func (p *GateClientPackEncrypt) DecodePacket(data []byte, pack *gatepack.GateClientPack) ecode.IErrCode {
	// pack := &gatepack.GateClientPack{}
	if pack == nil {
		return define.NilError
	}
	if err := pack.Deserialize(data); err != nil {
		return define.UnserializeError.SetError(err)
	}

	if int(pack.Length) > p.MaxPacketSize {
		return define.TooLargeError
	}

	if pack.Ver < 100 {
		return define.PackVerError
	}

	if pack.CmdId < 1 {
		return define.CmdidFieldError
	}

	p.handleDecrypt(pack)

	if err := p.handleUncompress(pack); err != nil {
		return define.UncompressError.SetError(err)
	}

	return nil
}

// type GateProtocol struct {
// 	GatePackEncrypt
// }

type GateCodecBase struct {
	*GateClientPackEncrypt
	id      uint32
	conn    net.Conn
	reader  *bufio.Reader
	headBuf []byte
	headDat [TcpGatePackageHeadLength]byte
}

// ////////
type TlsGateCodec struct {
	GateCodecBase
	realIp string
}

func (p *TlsGateCodec) ClearSendChan(sendChan <-chan interface{}) {
	for msg := range sendChan {
		p.Free(*(msg.(*[]byte)))
	}
}

func (p *TlsGateCodec) SocketID() uint32 {
	return p.id
}

func (p *TlsGateCodec) ClientAddr() string {
	return p.conn.RemoteAddr().String()
}

func (p *TlsGateCodec) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}

func (p *TlsGateCodec) Receive() (interface{}, error) {
	p.MemSet(p.headBuf, 0)
	if _, err := io.ReadFull(p.reader, p.headBuf); err != nil {
		return nil, err
	}
	length := int(binary.LittleEndian.Uint16(p.headBuf[0:]))
	if length > p.MaxPacketSize {
		return nil, define.TooLargeErr
	}
	buff := p.Alloc(p.HeadSize + length)
	p.MemSet(buff, 0)
	copy(buff, p.headBuf)
	if _, err := io.ReadFull(p.reader, buff[p.HeadSize:]); err != nil {
		p.Free(buff)
		return nil, err
	}
	return &buff, nil
}

func (p *TlsGateCodec) Send(msg interface{}) error {
	buffer := *(msg.(*[]byte))
	_, err := p.conn.Write(buffer)
	p.Free(buffer)
	return err
}

func (p *TlsGateCodec) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// ////
type TcpGateCodec struct {
	GateCodecBase
	realIp string
}

func (p *TcpGateCodec) ClearSendChan(sendChan <-chan interface{}) {
	for msg := range sendChan {
		p.Free(*(msg.(*[]byte)))
	}
}

func (p *TcpGateCodec) SocketID() uint32 {
	return p.id
}

func (p *TcpGateCodec) ClientAddr() string {
	return p.conn.RemoteAddr().String()
}

func (p *TcpGateCodec) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}

func (p *TcpGateCodec) Receive() (interface{}, error) {
	p.MemSet(p.headBuf, 0)
	if _, err := io.ReadFull(p.reader, p.headBuf); err != nil {
		return nil, err
	}
	length := int(binary.LittleEndian.Uint16(p.headBuf[0:]))
	if length > p.MaxPacketSize {
		return nil, define.TooLargeErr
	}
	buff := p.Alloc(p.HeadSize + length)
	p.MemSet(buff, 0)
	copy(buff, p.headBuf)

	if _, err := io.ReadFull(p.reader, buff[p.HeadSize:]); err != nil {
		p.Free(buff)
		return nil, err
	}
	// logs.Debug("head %v, receive %v",  p.headBuf, buff[p.HeadSize:])
	return &buff, nil
}

func (p *TcpGateCodec) Send(msg interface{}) error {
	buffer := *(msg.(*[]byte))
	_, err := p.conn.Write(buffer)
	p.Free(buffer)
	return err
}

func (p *TcpGateCodec) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// *********************************************
type WsGateCodec struct {
	*GateClientPackEncrypt
	id      uint32
	conn    *websocket.Conn
	headBuf []byte
	headDat [TcpGatePackageHeadLength]byte
	realIp  string
}

func (p *WsGateCodec) Receive() (interface{}, error) {
	p.MemSet(p.headBuf, 0)

	_, r, err := p.conn.NextReader()
	if err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(r, p.headBuf); err != nil {
		return nil, err
	}
	length := int(binary.LittleEndian.Uint16(p.headBuf[0:4]))
	if length > p.MaxPacketSize {
		return nil, define.TooLargeErr
	}

	buffer := p.Alloc(p.HeadSize + length)
	p.MemSet(buffer, 0)
	copy(buffer, p.headBuf)
	if _, err := io.ReadFull(r, buffer[p.HeadSize:]); err != nil {
		p.Free(buffer)
		return nil, err
	}
	return &buffer, nil
}

func (p *WsGateCodec) Send(msg interface{}) error {
	buffer := *(msg.(*[]byte))
	err := p.conn.WriteMessage(websocket.BinaryMessage, buffer) // p.conn.Write(buffer)
	p.Free(buffer)
	return err
}

func (p *WsGateCodec) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *WsGateCodec) SocketID() uint32 {
	return p.id
}

func (p *WsGateCodec) ClientAddr() string {
	return p.conn.RemoteAddr().String()
}

func (p *WsGateCodec) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}
