package dal

import (
	"fmt"

	"github.com/zhikongming/stock/biz/config"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

const (
	StatusEnabled  = 1
	StatusDisabled = 0
)

var (
	db *gorm.DB
)

func InitMysql(conf *config.Config) {
	// 初始化数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.DB.User,
		conf.DB.Password,
		conf.DB.Host,
		conf.DB.Port,
		conf.DB.DBName,
	)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}
}

func GetDB() *gorm.DB {
	return db
}
