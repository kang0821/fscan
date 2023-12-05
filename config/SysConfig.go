package config

var Config SysConfig

type SysConfig struct {
	Port  int   `yaml:"port"`
	Scan  Scan  `yaml:"scan"`
	Minio Minio `yaml:"minio"`
	Redis Redis `yaml:"redis"`
	Mysql Mysql `yaml:"mysql"`
}

type Scan struct {
	// 漏洞模板同步策略（ONCE启动时同步 < INTERVAL定时同步 < ALWAYS实时同步， 后一个总是包含前面所有的策略。如配置ALWAYS时，则会同时启用全部三种策略）
	TemplateSyncStrategy TemplateSyncStrategy
}

type Minio struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	Secure          string `yaml:"secure"`
	Bucket          string `yaml:"bucket"`
	Path            string `yaml:"path"`
	FileUrlPrefix   string `yaml:"fileUrlPrefix"`
}

type Redis struct {
	Nodes    []string `yaml:"nodes"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
}

type Mysql struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Database     string `yaml:"database"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
}

type TemplateSyncStrategy string

const (
	ONCE     TemplateSyncStrategy = "ONCE"
	INTERVAL TemplateSyncStrategy = "INTERVAL"
	ALWAYS   TemplateSyncStrategy = "ALWAYS"
)
