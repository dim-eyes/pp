package base

import (
	"encoding/binary"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	CreateConnChan  = make(chan *ITcpConn, 128)        //连接新建通道
	CloseConnChan   = make(chan *ITcpConn, 128)        //连接断开通道
	RecieveConnChan = make(chan MessageDataChan, 1024) //消息通道
	//CreateInnerConnChan  = make(chan *ITcpConn, 10)         //连接新建通道
	//CloseInnerConnChan   = make(chan *ITcpConn, 10)         //连接断开通道
	//RecieveInnerConnChan = make(chan MessageDataChan, 1024) //消息通道
)

type MessageDataChan struct {
	MsgID uint32
	Data  []byte
	Conn  *ITcpConn
}
type MessageData struct {
	MsgID uint32
	Data  []byte
}
type ITcpConn struct {
	ID             uint64    //连接唯一ID
	Addr           string    //客户端的连接地址
	EnterTime      time.Time //连接创建时间
	ServerType     int       // InnerServer 和 OutServer
	Conn           *net.Conn //底层连接
	CreateConnFlag chan int  //链接建立成功通道
}

func (c *ITcpConn) String() string {
	return c.Addr + ", CID:" + strconv.FormatUint(c.ID, 10) + ", Enter At:" +
		c.EnterTime.Format("2006-01-02 15:04:05+8000")
}

func (c *ITcpConn) SetConn(conn *net.Conn) {
	c.Conn = conn
}

func (c *ITcpConn) SendMsg(data MessageData) {
	(*c.Conn).Write(packMsg(&data))
}

const (
	LittleEndian = 1
	BigEndian    = 2

	ByteOrder = LittleEndian
)

func packMsg(data *MessageData) []byte {
	var msgLen int = 4 + len(data.Data)
	var packData []byte = make([]byte, 0)
	var msgHeadData []byte = make([]byte, 4)
	var msgIdData []byte = make([]byte, 4)
	if ByteOrder == LittleEndian {
		binary.LittleEndian.PutUint32(msgHeadData, uint32(msgLen))
		binary.LittleEndian.PutUint32(msgIdData, data.MsgID)
	} else {
		binary.BigEndian.PutUint32(msgHeadData, uint32(msgLen))
		binary.BigEndian.PutUint32(msgIdData, data.MsgID)
	}
	packData = append(packData, msgHeadData...)
	packData = append(packData, msgIdData...)
	packData = append(packData, data.Data...)
	return packData
}

func (c *ITcpConn) Close() {
	if c.Conn == nil {
		return
	}
	(*c.Conn).Close()
}

// 生成链接ID
var (
	globalID uint64
	idLocker sync.Mutex
)

const Uint64Max = ^uint64(0)

func GenConnID() uint64 {
	idLocker.Lock()
	defer idLocker.Unlock()

	if globalID >= Uint64Max {
		globalID = 0
	} else {
		globalID++
	}
	return globalID
}
