package dal

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/zhikongming/stock/biz/model"
	"gorm.io/driver/mysql"
)

var (
	db *gorm.DB
)

func InitMysql(conf *model.Config) {
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
