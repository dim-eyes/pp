package proto

// -----------------------协议号定义-------------------------------

const (
	ClientGateBeatHeart          = 10000 // 心跳
	InnerServerRegister          = 11003 // 游戏服务器注册
	ServerToServer               = 11004 // 网关转发服务器消息到其他服务器处理
	ClientToServer               = 11005 // 客户端转发给其他服务器
	GrpcToServer                 = 11006 // Grpc的消息转发
	ServerToClient               = 11007 // 服务器发送给客户端的消息
	ServerToGrpc                 = 11008 // 服务器回消息给grpc客户端
	ProtoNotifyServerState       = 11013 // 通知其他所有服务器该服务器状态变化
	ProtoServerLoadConfig        = 11014 // 通知各个游戏服务器加载配置
	ProtoStopServer              = 11016 // 服务器停服
	ProtoStopTargetServer        = 11017 // 停指定服务器
	ProtoNotifyInnerConnState    = 11020 // 服务器通知网关消息,服务处于维护中
	ProtoNotifyInnerConnCanClose = 11021 // 网关消息回复可以关闭
)
