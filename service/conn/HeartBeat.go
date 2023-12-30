package conn

// GateClientHeartBeatHandler 网关心跳回包
func GateClientHeartBeatHandler(conn *GateClient, serverID int, msgID uint32, data []byte) {
	conn.GetHeartBeatMsg()
}

func GateClientCloseRespHandler(conn *GateClient, serverID int, msgID uint32, data []byte) {
	GetGateClientMgr().StopConnCount++
}
