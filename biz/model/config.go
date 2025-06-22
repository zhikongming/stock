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
