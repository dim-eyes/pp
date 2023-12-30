package service

import (
	"pp/proto"
	gate "pp/service/conn"
	"sync"
)

var (
	mgr                *MsgHandlerMgr
	roomMsgHandlerOnce sync.Once
)

type HandlerMsg func(conn *gate.GateClient, userID int, msgID uint32, data []byte)

func GetMsgHandlerMgr() *MsgHandlerMgr {
	roomMsgHandlerOnce.Do(func() {
		if mgr == nil {
			mgr = &MsgHandlerMgr{make(map[uint32]HandlerMsg)}
		}
	})

	return mgr
}

type MsgHandlerMgr struct {
	msgHandlerFunc map[uint32]HandlerMsg
}

func (m *MsgHandlerMgr) RegisterMsgHandlerFunc(msgID uint32, doHandler func(conn *gate.GateClient, userID int, msgID uint32, data []byte)) {
	m.msgHandlerFunc[msgID] = doHandler
}

func (m *MsgHandlerMgr) GetMsgHandler(msgID uint32) (HandlerMsg, bool) {
	handler, ok := m.msgHandlerFunc[msgID]
	if !ok {
		return nil, false
	}
	return handler, true
}

// Init 初始化网关消息处理
func (m *MsgHandlerMgr) init() bool {
	m.RegisterMsgHandlerFunc(proto.ClientGateBeatHeart, gate.GateClientHeartBeatHandler) // 心跳处理

	return true
}
