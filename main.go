package main

import (
	"github.com/shadow1ng/fscan/client"
	"github.com/shadow1ng/fscan/config"
	yml_config "github.com/shadow1ng/fscan/util"
)

func main() {
	//_ = routers.InitApiRouter().Run(":" + strconv.Itoa(config.Config.Port))
	//client.Test()
	yml_config.CreateYamlFactory("config.yml", &config.Config)
	//println(config.Config.Minio.FileUrlPrefix)
	//upload := client.Upload("E:\\test.png")
	//println(upload)

	//client.InitMinio(config.Config.Minio)
	//client.InitRedis(config.Config.Redis)
	//client.Tets()

	client.InitMysql(config.Config.Mysql)
	client.Test()
}
