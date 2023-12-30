package mysql

import (
	"fmt"
	"pp/log"
	"sync"

	glog "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"gorm.io/gorm"

	// init mysql driver
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
)

type DbType uint32

const (
	ActivityMysql DbType = 1
	VictoryMsql   DbType = 2
)

var (
	client            *DbMgr
	clientOnce        sync.Once
	logger            = log.GetLogger()
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

func Client() *DbMgr {
	return client
}

type DbMgr struct {
	clients map[DbType]*gorm.DB
	mutex   sync.RWMutex
}

func NewDbMgr() *DbMgr {
	clientOnce.Do(func() {
		client = &DbMgr{
			clients: make(map[DbType]*gorm.DB),
			mutex:   sync.RWMutex{},
		}
	})
	return client
}

// Db 获取gorm.db
func (d *DbMgr) Db(Type DbType) (*gorm.DB, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	db, ok := d.clients[Type]
	if db == nil {
		return db, false
	}
	return db, ok
}

// SetDb 添加redisClient
func (d *DbMgr) SetDb(Type DbType, db *gorm.DB) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.clients[Type] = db
}

// NewDB 初始化mysql
func NewDB(addr, userName, pwd, dbName string, opts ...Option) *gorm.DB {
	options := options{
		dblog: false,
	}
	for _, o := range opts {
		o(&options)
	}
	url := generateDBUrl(userName, pwd, addr, dbName)
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "victory_", // 表名前缀
			SingularTable: true,       // 关闭复数表明
		},
		SkipDefaultTransaction: true, // 禁用默认事务
	})
	if !options.dblog {
		db.Logger = glog.Default.LogMode(glog.Silent)
	} else {
		db.Logger = glog.Default.LogMode(glog.Info)
	}
	if err != nil {
		fmt.Println("failed opening connection to mysql: ", err)
		logger.Fatal("failed opening connection to mysql: %v", err)
	}
	sqlDB, _ := db.DB()
	if options.maxIdleConn > 0 {
		sqlDB.SetMaxIdleConns(options.maxIdleConn)
	}
	if options.maxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(options.maxLifetime)
	}
	if options.maxOpenConn > 0 {
		sqlDB.SetMaxOpenConns(options.maxOpenConn)
	}
	return db
}

// generateDBUrl 生成mysql的链接地址
func generateDBUrl(userName, pwd, addr, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", userName, pwd, addr, dbName)
}
