package client

import (
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"github.com/shadow1ng/fscan/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

type MysqlContext struct {
	MysqlClient *gorm.DB
}

func InitMysql(mysqlConfig config.Mysql) {
	var err error
	common.Context.Mysql.MysqlClient, err = gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database),
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 禁用表名复数
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	// ----------------------------数据库连接池----------------------------
	sqlDB, err := common.Context.Mysql.MysqlClient.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
}
