package util

import (
	"github.com/spf13/viper"
	"log"
)

// CreateYamlFactory 创建一个yaml配置文件工厂
func CreateYamlFactory(fileName string, config any) *viper.Viper {
	yamlConfig := viper.New()
	yamlConfig.SetConfigFile(fileName)
	if err := yamlConfig.ReadInConfig(); err != nil {
		log.Fatal("配置文件初始化失败", err.Error())
	}
	if err := yamlConfig.Unmarshal(config); err != nil {
		log.Fatal("配置文件解析失败", err.Error())
	}
	return yamlConfig
}
