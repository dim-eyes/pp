package service

import (
	gate "pp/service/conn"
	"time"
	"unsafe"
)

// StartMessageProcess 先在一个协程中处理，，消息处理在开个协程单独处理
func StartMessageProcess() {
	// 消息链接管理
	// 收到玩家建立链接断开链接和消息处理
	logger.Info("start conn message and close message process")
	for {
		select {
		case msg := <-gate.MessageDataChan:
			logger.Debug("handler msg start, msgID:", msg.Data.MsgID)
			go ProcessOneMessage(msg.Client, msg.Data.MsgID, msg.Data.Data)
		}
	}
}

// ProcessOneMessage 消息处理
func ProcessOneMessage(conn *gate.GateClient, msgID uint32, data []byte) {
	handlerMsgID := uint32(0)
	var handlerData []byte
	var userID int

	handlerMsgID = msgID
	handlerData = data
	userID = conn.ServerID

	msgMgr := GetMsgHandlerMgr()
	handler, ok := msgMgr.GetMsgHandler(handlerMsgID)
	if !ok {
		logger.Debug("handler msg can not find, serverID:", conn.ServerID, ",userID:", userID, ",msgID:", handlerMsgID, ",data:", *(*string)(unsafe.Pointer(&handlerData)))
		return
	}
	// 调用函数处理
	now := time.Now().UnixNano()
	handler(conn, userID, handlerMsgID, handlerData)
	logger.Info("handler msg, serverID:", conn.ServerID, ",userID:", userID, ",duration:", time.Now().UnixNano()-now, ",msgID:", handlerMsgID, ",data:", *(*string)(unsafe.Pointer(&handlerData)))
}
