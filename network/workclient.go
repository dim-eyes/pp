package network

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"pp/log"
	"pp/network/base"
	"strconv"
	"time"
)

var (
	maxMsgLen uint32 = 100 * 1024 * 1024
	minMsgLen uint32 = 8
	lenMsgLen        = 4
	logger           = log.GetLogger()
)

func GetConnect(addr string) (*NetClient, error) {
	client, err := net.DialTimeout("tcp", addr, time.Second*20)
	if err != nil {
		logger.Error("Connect failed," + addr)
		return nil, err
	}
	return &NetClient{Conn: &client}, nil
}

type NetClient struct {
	Conn *net.Conn
}

func (this *NetClient) Close() {
	if this.Conn != nil {
		(*this.Conn).Close()
	}
}

func (this *NetClient) SendMsg(msgID uint32, data []byte) {
	_, err := (*this.Conn).Write(this.PackMsg(msgID, data))
	if err != nil {
		logger.Error("send msg failed,", msgID, data, err.Error())
		return
	}
}

func (this *NetClient) PackMsg(msgID uint32, data []byte) []byte {
	var msgLen int = 4 + len(data)
	var packData []byte = make([]byte, 0)
	var msgHeadData []byte = make([]byte, 4)
	var msgIdData []byte = make([]byte, 4)
	if base.ByteOrder == base.LittleEndian {
		binary.LittleEndian.PutUint32(msgHeadData, uint32(msgLen))
		binary.LittleEndian.PutUint32(msgIdData, msgID)
	} else {
		binary.BigEndian.PutUint32(msgHeadData, uint32(msgLen))
		binary.BigEndian.PutUint32(msgIdData, msgID)
	}
	packData = append(packData, msgHeadData...)
	packData = append(packData, msgIdData...)
	packData = append(packData, data...)
	return packData
}

// TCP消息解析
func (this *NetClient) ReadOnePacketMsg(conn net.Conn) ([]byte, error, uint32) {
	//var b [4]byte
	bufMsgLen := make([]byte, 4)

	// read len
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return nil, err, 0
	}

	bufMsgId := make([]byte, 4)
	// read msg id
	if _, err := io.ReadFull(conn, bufMsgId); err != nil {
		return nil, err, 0
	}

	// parse len
	var msgLen uint32
	var msgId uint32
	switch lenMsgLen {
	case 1:
		msgLen = uint32(bufMsgLen[0])
	case 2:
		if base.ByteOrder == base.LittleEndian {
			msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
		} else {
			msgLen = uint32(binary.BigEndian.Uint16(bufMsgLen))
		}
	case 4:
		if base.ByteOrder == base.LittleEndian {
			msgLen = binary.LittleEndian.Uint32(bufMsgLen)
			msgId = binary.LittleEndian.Uint32(bufMsgId)
		} else {
			msgLen = binary.BigEndian.Uint32(bufMsgLen)
			msgId = binary.BigEndian.Uint32(bufMsgId)
		}
	}

	// check len
	if msgLen > maxMsgLen {
		return nil, errors.New("message too long" + strconv.Itoa(int(msgLen))), 0
	} else if msgLen < minMsgLen {
		return nil, errors.New("message too short"), 0
	}

	// data
	msgData := make([]byte, msgLen-4)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err, 0
	}

	return msgData, nil, msgId
}
