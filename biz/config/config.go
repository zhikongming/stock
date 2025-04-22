package config

import (
	"fmt"
	"os"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"gopkg.in/yaml.v3"
)

var conf *model.Config

func InitConfig() {
	config_path := "./config.yaml"
	data, err := os.ReadFile(config_path)
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败: %v", err))
	}

	if err := yaml.Unmarshal(data, &conf); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}

	// 初始化配置
	dal.InitMysql(conf)
}

func GetConfig() *model.Config {
	return conf
}
