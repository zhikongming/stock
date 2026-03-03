package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type ReplaceConfig struct {
	OriginDomain   string `yaml:"origin_domain"`
	ReplacedDomain string `yaml:"replaced_domain"`
}

type Config struct {
	// 配置项
	DB      *DBConfig      `yaml:"DB"`
	Server  *ServerConfig  `yaml:"Server"`
	Replace *ReplaceConfig `yaml:"Replace"`
	Lark    *LarkConfig    `yaml:"Lark"`
}

type LarkConfig struct {
	AppID         string `yaml:"app_id"`
	AppSecret     string `yaml:"app_secret"`
	TestReceiveID string `yaml:"test_receive_id"`
	GroupRobotURL string `yaml:"group_robot_url"`
}

var conf *Config

func InitConfig() {
	config_path := "./config.yaml"
	data, err := os.ReadFile(config_path)
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败: %v", err))
	}

	if err := yaml.Unmarshal(data, &conf); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}

	// 初始化 Lark 配置
	larkConfigPath := "./lark.yaml"
	larkData, err := os.ReadFile(larkConfigPath)
	if err != nil {
		return
	}
	if err := yaml.Unmarshal(larkData, &conf); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}
}

func GetConfig() *Config {
	return conf
}

func GetLarkConfig() *LarkConfig {
	return conf.Lark
}

func GetLocalHost() string {
	if conf.Replace == nil {
		return "http://localhost:6789"
	}
	return fmt.Sprintf("http://%s", conf.Replace.ReplacedDomain)
}
