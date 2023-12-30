package conn

import (
	"pp/db/redis"
	"pp/service/constant"
	"strconv"
)

// SendMsgToClient 通过userID发送给客户端
func SendMsgToClient(userID int, msgID uint32, data string) {
	if userID == 0 {
		logger.Warn("SendMsgToClient userID is zero,msgId:", msgID, ",data:", data)
	}
	pRedisMgr := redis.GetInstance()
	pRedis, index := pRedisMgr.GetRedisClientByType(redis.RedisTypePlayer)
	if index == 0 {
		logger.Error("can not find redis info,userID:", userID, msgID)
		return
	}
	gateIDStr, err := pRedis.HGet(constant.GetPlayerStatusKey(userID), "GateID")
	gateID, err := strconv.Atoi(gateIDStr)
	if err != nil {
		logger.Error("SendMsgToClient Atoi error:", err, ",gateIDStr:", gateIDStr)
		return
	}
	if gateID == 0 {
		logger.Info("SendMsgToClient user gateId is zero")
		return
	}
	client, ok := GetGateClientMgr().GetClient(gateID)
	if !ok {
		logger.Warn("SendMsgToClient GetClient gate not exist,gateID:", gateID)
		return
	}
	client.SendMsgToClient(userID, msgID, data)
	return
}
