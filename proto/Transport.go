package proto

// -----------------服务器与服务器,服务器与客户端,客户端与服务器消息传输结构体---------------------

// ClientToServerMsgProto 客户端消息，网关转发其他服务器
// MsgID: 10005
type ClientToServerMsgProto struct {
	UserID     int    // 玩家ID
	ServerID   int    // 本服务器的ServerID
	ServerType int    // 服务器类型
	MsgID      uint32 // 消息类型
	Data       string // 具体协议内容
}

// ServerToClientMsg 消息发送给客户端
type ServerToClientMsg struct { // msgID: 10006
	UserID int    // 玩家ID
	MsgID  uint32 // 消息类型
	Data   string // 具体协议内容
}

// GrpcToServerMsg GRPC ---> other server
type GrpcToServerMsg struct { //
	ConnID     int    // 链接ID
	ServerID   int    // 服务器的ServerID
	ServerType int    // 服务器类型
	MsgID      uint32 // 消息ID
	Data       []byte // 数据封装
}

// ServerToGrpcMsg Server ----> GRPC
type ServerToGrpcMsg struct {
	ConnID int    // 链接ID
	MsgID  uint32 // 消息ID
	Data   string // 数据封装
}

// ServerToServerMsg  Server --- > Server 转发单个服务器
type ServerToServerMsg struct {
	TargetServerID   int    `json:"targetserverid"`   // 服务器的ServerID
	TargetServerType int    `json:"targetservertype"` // 服务器类型
	ServerID         int    `json:"serverid"`         // 发送者的ServerID
	ServerType       int    `json:"servertype"`       // 发送者的ServerType
	MsgID            uint32 `json:"msgid"`            // 消息ID
	Data             string `json:"data"`             // 数据封装
}

// ServerToAllServerMsg server ----> allServer,转发给所有服务器类型等于serverType的服务器
type ServerToAllServerMsg struct {
	TargetServerType int    `json:"targetservertype"` // 服务器类型
	ServerID         int    `json:"serverid"`         // 发送者的ServerID
	ServerType       int    `json:"servertype"`       // 发送者的ServerType
	MsgID            uint32 `json:"msgid"`            // 消息ID
	Data             string `json:"data"`             // 数据封装
}

// RegisterServerInfo 服务注册结构体
type RegisterServerInfo struct {
	ServerID   int    `json:"id"`   // 服务ID
	ServerType int    `json:"type"` // 服务类型
	ServerName string `json:"name"` // 服务名称
}

// NotifyServerState ProtoNotifyServerState = 11013 //通知其他所有服务器该服务器状态变化
type NotifyServerState struct {
	State      int // 服务器状态
	ServerType int // 服务类型
	ServerID   int // 服务器的ServerID
}

// NotifyServerConfigUpdate ProtoServerLoadConfig = 11014 //通知各个游戏服务器加载配置
type NotifyServerConfigUpdate struct {
	UpdateKey []string `json:"updatekey"` // 更新配置的key
}

// NotifyStopTargetServer 后台指定服务器停服 ProtoStopTargetServer = 11017
type NotifyStopTargetServer struct {
	StopFlag   int `json:"stop"`      // 1：停服 0：开服
	ServerType int `json:"type"`      // 服务器类型： 3：房间服 4：游戏服
	ServerID   int `json:"server_id"` // 停指定服务器
	GameID     int `json:"game_id"`   // 停指定玩法
}

// NotifyGateAddOrRemove 增加网关和删除网关 ProtoAddOrRemoveGate = 11018
type NotifyGateAddOrRemove struct {
	Type     string `json:"type"`      // “Add" "Remove"
	Addr1    string `json:"addr1"`     // 玩法游戏服务器不用此字段
	Addr2    string `json:"addr2"`     // 玩家游戏服用
	ServerID int    `json:"server_id"` // 添加的网关的serverID
}
