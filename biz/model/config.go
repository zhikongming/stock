package model

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

type Config struct {
	// 配置项
	DB     *DBConfig     `yaml:"DB"`
	Server *ServerConfig `yaml:"Server"`
}
