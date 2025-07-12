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
}

func GetConfig() *Config {
	return conf
}
