package conn

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"pp/common"
	"pp/config"
	"pp/log"
	"pp/network"
	"pp/network/base"
	"pp/proto"
	"sync"
	"time"
)

type TcpClientMessageChan struct {
	Client *GateClient
	Data   base.MessageData
}

var (
	MessageDataChan chan TcpClientMessageChan = make(chan TcpClientMessageChan, 10000)
	logger                                    = log.GetLogger()
)

type GateClient struct {
	ServerID     int                // 要连接服务器ServerID
	ServerType   int                // 要连接服务器的ServerType
	Addr         string             // 连接地址
	client       *network.NetClient // 底层socket连接
	timeoutCount int                // 心跳超时次数
	timestamp    int64              // 心跳开始时间
}

func (g *GateClient) String() string {
	str, err := json.Marshal(g)
	if err != nil {
		return ""
	}
	return "GateClient:" + string(str)
}

// Start 启动和网关链接的客户端
func (g *GateClient) Start() {
	for {
		// 如果链接断开，此处保持1s重连
		client, err := network.GetConnect(g.Addr)
		if err == nil {
			g.client = client
			// 连接建立后发送服务注册消息
			g.RegisterServerToGate()
			GetGateClientMgr().AddClient(g)
		} else {
			time.Sleep(time.Second)
			continue
		}

		for {
			// 等待接收消息
			data, recvErr, msgID := client.ReadOnePacketMsg(*client.Conn)
			if recvErr != nil {
				logger.Error("GateClient connect error:", recvErr.Error())
				client.Close()
				break
			}
			tcpClient := &base.ITcpConn{ID: base.GenConnID(), Addr: g.Addr, EnterTime: time.Now(), ServerType: g.ServerType}
			tcpClient.SetConn(client.Conn)
			MessageDataChan <- TcpClientMessageChan{Data: base.MessageData{MsgID: msgID, Data: data}, Client: g}
		}
		// 删除该链接
		gateMgr := GetGateClientMgr()
		gateMgr.RemoveClient(g.ServerID)
		time.Sleep(time.Second)
	}
}

// RegisterServerToGate 向网关注册服务
func (g *GateClient) RegisterServerToGate() {
	appConfig := config.NewAppConfig().GetConfig()
	data := &proto.RegisterServerInfo{ServerID: appConfig.ServerID, ServerType: appConfig.ServerType, ServerName: appConfig.ServerName}
	msg, err := json.Marshal(data)
	if err != nil {
		logger.Error("register server to gate failed,", string(msg))
		return
	}
	g.client.SendMsg(proto.InnerServerRegister, msg)
}

// SendMsgToClient 发送消息给客户端
func (g *GateClient) SendMsgToClient(userID int, msgID uint32, data string) {
	var msg proto.ServerToClientMsg
	msg.UserID = userID
	msg.MsgID = msgID
	msg.Data = data
	sendData, err := json.Marshal(&msg)
	if err != nil {
		logger.Error("SendMsgToClient,data format error,", err.Error())
		return
	}
	g.client.SendMsg(proto.ServerToClient, sendData)
}

func (g *GateClient) CheckHeartBeatTimeout() {
	now := time.Now().Unix()
	if g.timestamp != 0 && g.timestamp+5 < now {
		g.timeoutCount++
		g.timestamp = now
		if g.timeoutCount >= 3 {
			g.timeoutCount = 0
			g.timestamp = 0
			g.client.Close()
			logger.Error("CheckHeartBeatTimeout need reconnect gate")
		}
	}
}

// GetHeartBeatMsg  收到心跳消息
func (g *GateClient) GetHeartBeatMsg() {
	g.timestamp = 0 // 设置心跳时间
	g.timeoutCount = 0
}

func (g *GateClient) SendHeartBeatMsg() {
	g.client.SendMsg(proto.ClientGateBeatHeart, common.Str2bytes(`{"msgID":10000}`))
	g.timestamp = time.Now().Unix()
}

// SendStopServerMsg needRet: 是否需要返回
func (g *GateClient) SendStopServerMsg(needRet int) {
	g.client.SendMsg(proto.ProtoNotifyInnerConnState, common.Str2bytes(fmt.Sprintf(`{"isRet":%v}`, needRet)))
}

// SendMsgToServer 发送消息给其他服务器
func (g *GateClient) SendMsgToServer(serverID, serverType int, msgID uint32, data []byte) {
	var msg proto.ServerToServerMsg
	msg.TargetServerID = serverID
	msg.TargetServerType = serverType
	msg.ServerID = config.NewAppConfig().GetConfig().ServerID
	msg.ServerType = config.NewAppConfig().GetConfig().ServerType
	msg.MsgID = msgID
	msg.Data = string(data)
	sendData, err := json.Marshal(&msg)
	if err != nil {
		return
	}
	g.client.SendMsg(proto.ServerToServer, sendData)
	logger.Info(fmt.Sprintf(" gateConn SendMsgToServer, %v,%v,%v,%v", g.ServerID, serverID, serverType, msgID))
}

// SendMsgToGate 发消息到网关
func (g *GateClient) SendMsgToGate(msgID uint32, data []byte) {
	g.client.SendMsg(msgID, data)
}

// SendMsgToGrpc 发送消息给grpc客户端
func (g *GateClient) SendMsgToGrpc(connID int, msgID uint32, data []byte) {
	var msg proto.ServerToGrpcMsg
	msg.ConnID = connID
	msg.MsgID = msgID
	msg.Data = string(data)
	sendData, err := json.Marshal(&msg)
	if err != nil {
		return
	}
	g.client.SendMsg(proto.ServerToGrpc, sendData)
}

var (
	gateClientMgrOnce sync.Once
	gateClientMgr     *GateClientMgr
)

func GetGateClientMgr() *GateClientMgr {
	gateClientMgrOnce.Do(func() {
		if gateClientMgr == nil {
			gateClientMgr = &GateClientMgr{}
		}
	})
	return gateClientMgr
}

// GateClientMgr 客户端管理
type GateClientMgr struct {
	GateClientMap sync.Map
	Count         int
	timerCount    int64 // 计时器
	StopConnCount int
}

// AddClient 建立一个连接
func (g *GateClientMgr) AddClient(client *GateClient) {
	g.GateClientMap.Store(client.ServerID, client)
	g.Count++
	logger.Debug("GateClientMgr:AddClient, serverID:", client.ServerID)
}

// GetClient 精确定位一个client
func (g *GateClientMgr) GetClient(serverID int) (*GateClient, bool) {
	client, ok := g.GateClientMap.Load(serverID)
	if !ok {
		return nil, false
	}
	return client.(*GateClient), true
}

// RemoveClient 删除客户端
func (g *GateClientMgr) RemoveClient(serverID int) bool {
	_, load := g.GateClientMap.LoadAndDelete(serverID)
	if load {
		if g.Count > 0 {
			g.Count--
		}
	}
	logger.Debug("GateClientMgr:RemoveClient, serverID:", serverID)
	return true
}

// RandOneClient 随机找一个网关发送消息
func (g *GateClientMgr) RandOneClient() (*GateClient, bool) {
	if g.Count == 0 {
		return nil, false
	}
	randCount := rand.Intn(g.Count)
	count := 0
	var client *GateClient
	g.GateClientMap.Range(func(key, value interface{}) bool {
		count++
		if count >= randCount {
			gateClient, ok := value.(*GateClient)
			if ok {
				client = gateClient
				return false
			} else {
				return true
			}
		}
		return true
	})

	if client == nil {
		return nil, false
	}
	return client, true
}

// SendHeartBeat 发送心跳消息
func (g *GateClientMgr) SendHeartBeat() {
	g.GateClientMap.Range(func(key, value interface{}) bool {
		gateClient, ok := value.(*GateClient)
		if ok {
			gateClient.SendHeartBeatMsg()
		}
		return true
	})
}

// CheckHeartBeatTimeout 心跳超时检测
func (g *GateClientMgr) CheckHeartBeatTimeout() {
	g.GateClientMap.Range(func(key, value interface{}) bool {
		gateClient, ok := value.(*GateClient)
		if ok {
			gateClient.CheckHeartBeatTimeout()
		}
		return true
	})
}

func (g *GateClientMgr) Timer1s() {
	g.CheckHeartBeatTimeout()
	g.timerCount++
	if g.timerCount%60 == 0 {
		g.Time1min()
	}
}

// 1分钟定时器
func (g *GateClientMgr) Time1min() {
	g.SendHeartBeat()
}

func (g *GateClientMgr) SendStopServerMsg(needRet int) {
	g.GateClientMap.Range(func(key, value interface{}) bool {
		gateClient, ok := value.(*GateClient)
		if ok {
			gateClient.SendStopServerMsg(needRet)
		}
		return true
	})
}

// 广播给所有网关
func (this *GateClientMgr) BroadcastAllGate(msgID uint32, data []byte) {
	this.GateClientMap.Range(func(key, value interface{}) bool {
		gateClient, ok := value.(*GateClient)
		if ok {
			gateClient.SendMsgToClient(0, msgID, string(data))
		}
		return true
	})
}
