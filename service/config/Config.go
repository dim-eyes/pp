package config

import (
	"pp/log"
)

// Configer  配置接口
type Configer interface {
	LoadConfig() bool
	ToString() string
}

var (
	appConfigMgr *appConfigManager
	logger       = log.GetLogger()
)

type appConfigManager struct {
	config map[string]*Configer
}

func NewAppConfigMgr() *appConfigManager {
	if appConfigMgr == nil {
		appConfigMgr = &appConfigManager{make(map[string]*Configer)}
		logger.Info("New app config mgr")
	}

	return appConfigMgr
}

// RegisterConfig 注册配置
func (a *appConfigManager) RegisterConfig(config *Configer) {
	a.config[(*config).ToString()] = config
}

// LoadAllConfig 加载所有配置
func (a *appConfigManager) LoadAllConfig() bool {
	for _, config := range a.config {
		if !(*config).LoadConfig() {
			logger.Error("load config " + (*config).ToString() + " failed!")
			return false
		}
		logger.Info("load config " + (*config).ToString() + " success")
	}
	return true
}

// ReloadConfig 重新加载配置
func (a *appConfigManager) ReloadConfig(update []string) bool {
	for _, configName := range update {
		config, ok := a.config[configName]
		if ok {
			if (*config).LoadConfig() {
				logger.Info("ReloadConfig " + configName + " success!")
			} else {
				logger.Error("ReloadConfig " + configName + " failed!")
				continue
			}
		}
	}
	return true
}

// Init 初始化配置
func (a *appConfigManager) Init() bool {
	return true
}
