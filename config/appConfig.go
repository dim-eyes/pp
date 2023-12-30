package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// AppConfig app.json配置读取
type AppConfig struct {
	appConfig AppConfigInfo
}

var appConfig *AppConfig

func NewAppConfig() *AppConfig {
	if appConfig == nil {
		appConfig = &AppConfig{}
	}
	return appConfig
}

type RedisConfig struct {
	RedisAddr string `json:"addr"`
	RedisType int    `json:"redistype"`
	Password  string `json:"password"`
}

type StartServerConfig struct {
	Type      int    `json:"type"`      // 启动的端口类型 1：TCP服务 2：websocket 3：GRPC服务 4：http服务
	OutAddr   string `json:"outaddr"`   // 对外开放地址
	InnerAddr string `json:"inneraddr"` // 对内开放地址
}

type ServersConfig struct {
	ServerID   int    `json:"serverid"`   // ServerID
	ServerType int    `json:"servertype"` // 服务类型
	Addr       string `json:"addr"`       // 连接的Server的IP和端口：ip:port
}

type MysqlConfig struct {
	Type     int    `json:"type"`
	Addr     string `json:"addr"`
	UserName string `json:"userName"`
	Pwd      string `json:"pwd"`
	DbName   string `json:"dbName"`
	Dblog    bool   `json:"dblog"`
}

type AppConfigInfo struct {
	ServerID          int                 `json:"serverid"`   // 服务器ID
	ServerType        int                 `json:"servertype"` // 服务器类型
	ServerName        string              `json:"servername"` // 服务器名称
	ServerPort        []StartServerConfig `json:"network"`    // 服务需要开启的端口信息
	ConnServersConfig []ServersConfig     `json:"servers"`    // 服务器需要连接的服务器信息
	RedisConfig       []RedisConfig       `json:"redis"`      // 服务器需要连接Redis的配置
	MysqlConfig       []*MysqlConfig      `json:"mysql"`      // mysql连接配置
	HttpUrlRoot       string              `json:"url"`        // 请求PHP的根地址
	Logger            string              `json:"logger"`     // 日志配置路径，包含日志文件名头部
	LoggerLevel       int                 `json:"level"`      // 日志级别
	LoggerFileMax     int64               `json:"logfilemax"` // 日志文件最大大小限制
	BiApiPath         string              `json:"biurl"`      // nginx打点api地址
}

func (c *AppConfig) LoadConfig() bool {
	path := "./app.json"
	//os.ReadFile
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("load config can not find " + c.ToString())
		return false
	}
	var tempConfig = AppConfigInfo{}
	jsonErr := json.Unmarshal(data, &tempConfig)
	if jsonErr != nil {
		fmt.Println("config file format error "+c.ToString()+" "+jsonErr.Error(), ",data:", string(data))
		return false
	}
	c.appConfig = tempConfig
	fmt.Println("app config load success," + string(data))
	return true
}

func (c *AppConfig) ToString() string {
	return "app.json"
}

func (a *AppConfig) GetConfig() AppConfigInfo {
	return a.appConfig
}
