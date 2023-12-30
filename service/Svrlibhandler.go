package service

import (
	"fmt"
	"math/rand"
	"os"
	"pp/config"
	"pp/db/mysql"
	"pp/service/timer"

	"pp/db/redis"
	"pp/log"
	serviceConfig "pp/service/config"
	gate "pp/service/conn"
	"strconv"
	"sync"
	"time"
)

var (
	svrlibHandler *Svrlibhandler
	createOnce    = sync.Once{}
	logger        = log.GetLogger()
)

func GetSvrlibhandler() *Svrlibhandler {
	createOnce.Do(func() {
		if svrlibHandler == nil {
			svrlibHandler = &Svrlibhandler{}
		}
	})
	return svrlibHandler
}

type Svrlibhandler struct {
}

func (s *Svrlibhandler) OnInit() bool {
	rand.Seed(time.Now().UnixNano())
	appConfig := config.NewAppConfig().GetConfig()
	// 建立redis客户端链接，方便后续调用
	redisMgr := redis.GetInstance()
	for _, redisInfo := range appConfig.RedisConfig {
		redisClient := &redis.RedisClient{ConnString: redisInfo.RedisAddr, Password: redisInfo.Password}
		ok := redisClient.ConnRedis()
		if !ok {
			fmt.Println("get Redis client failed")
			return false
		}
		redisMgr.AddRedisClientByType(redisInfo.RedisType, redisClient)
	}
	// 建立mysql客户端链接，方便后续调用
	mysqlMgr := mysql.NewDbMgr()
	for _, mysqlInfo := range appConfig.MysqlConfig {
		if mysqlInfo.DbName == "" {
			logger.Fatal("OnInit MysqlConfig DbName is null")
		}
		db := mysql.NewDB(mysqlInfo.Addr, mysqlInfo.UserName, mysqlInfo.Pwd, mysqlInfo.DbName,
			mysql.WithDbLog(mysqlInfo.Dblog),
			mysql.WithMaxIdleConn(25),
			mysql.WithMaxOpenConn(126),
			mysql.WithMaxLifetime(30*time.Minute))
		mysqlMgr.SetDb(mysql.DbType(mysqlInfo.Type), db)
	}

	// 注册服务器配置文件读取
	configMgr := serviceConfig.NewAppConfigMgr()
	if !configMgr.Init() {
		return false
	}

	// 初始化日志系统
	logger := log.GetLogger()
	if !logger.InitLogger() {
		return false
	}
	// 服务器启动读取所有配置
	if !configMgr.LoadAllConfig() {
		return false
	}

	//消息处理函数注册
	msgHandler := GetMsgHandlerMgr()
	if !msgHandler.init() {
		logger.Error("GetMsgHandlerMgr init failed")
		return false
	}

	// 读取需要连接的服务器数据
	for _, serverInfo := range appConfig.ConnServersConfig {
		client := &gate.GateClient{ServerID: serverInfo.ServerID, ServerType: serverInfo.ServerType, Addr: serverInfo.Addr}
		go client.Start()
	}
	// 启动定时器
	go timer.GetTickTimerMgr().Timer()
	logger.Info("OnInit success")
	return true
}

// 初始化全局基础数据，例如 房间号生成起始值
func (s *Svrlibhandler) InitBaseData() bool {
	pRedisMgr := redis.GetInstance()
	pRedis, index := pRedisMgr.GetRedisClientByType(redis.RedisTypePlayer)
	if index == 0 {
		return false
	}
	result, _ := pRedis.HGet("room:generate:roomid:incr", "RoomID")
	roomID, err := strconv.Atoi(result)
	if err != nil || roomID == 0 {
		pRedis.HSet("room:generate:roomid:incr", "RoomID", 150000)
	}
	return true
}

// 进程退出后需要处理相关逻辑
func (s *Svrlibhandler) OnQuit() {
	logger.Info("Service OnQuit Start, pid:", os.Getpid(), ", ServerName:", config.NewAppConfig().GetConfig().ServerName, "ServerID:", config.NewAppConfig().GetConfig().ServerID)

	//向网关广播服务器停服
	gate.GetGateClientMgr().SendStopServerMsg(1)

	go StartMessageProcess()

	//等待服务器处理完所有消息
	count := 0
	for {
		time.Sleep(time.Second)
		if len(gate.MessageDataChan) == 0 && gate.GetGateClientMgr().Count == gate.GetGateClientMgr().StopConnCount {
			break
		}
		count++
		if count > 3 {
			break
		}
	}
	// 进程退出的时候处理
	logger.Info("Service OnQuit End, pid:", os.Getpid(), ", ServerName:", config.NewAppConfig().GetConfig().ServerName, "ServerID:", config.NewAppConfig().GetConfig().ServerID)
}

func (s *Svrlibhandler) ReloadAppConfig() {
	appJson := config.NewAppConfig()

	connServerList := appJson.GetConfig().ConnServersConfig[:]
	redisList := appJson.GetConfig().RedisConfig[:]

	appJson.LoadConfig()
	//设置日志级别
	logger.SetLevel(appJson.GetConfig().LoggerLevel)
	logger.SetLogFileMax(appJson.GetConfig().LoggerFileMax)

	connServerListNew := appJson.GetConfig().ConnServersConfig
	redisListNew := appJson.GetConfig().RedisConfig

	for _, serverInfo := range connServerListNew {
		isFind := false
		for _, serverOldInfo := range connServerList {
			if serverInfo.ServerID == serverOldInfo.ServerID {
				isFind = true
				break
			}
		}
		if !isFind {
			client := &gate.GateClient{ServerID: serverInfo.ServerID, ServerType: serverInfo.ServerType, Addr: serverInfo.Addr}
			go client.Start()
			logger.Info("ReloadAppConfig add serverInfo,", client)
		}
	}

	redisMgr := redis.GetInstance()
	for _, redisInfo := range redisListNew {
		isFind := false
		for _, redisOldInfo := range redisList {
			if redisInfo.RedisType == redisOldInfo.RedisType {
				isFind = true
				break
			}
		}
		if !isFind {
			redisClient := &redis.RedisClient{ConnString: redisInfo.RedisAddr, Password: redisInfo.Password}
			ok := redisClient.ConnRedis()
			if !ok {
				fmt.Println("get Redis client failed")
				continue
			}
			redisMgr.AddRedisClientByType(redisInfo.RedisType, redisClient)
			logger.Info("ReloadAppConfig add redis server info,", redisClient)
		}
	}
}
