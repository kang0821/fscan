package main

import (
	"github.com/shadow1ng/fscan/WebScan"
	"github.com/shadow1ng/fscan/client"
	"github.com/shadow1ng/fscan/config"
	"github.com/shadow1ng/fscan/routers"
	"github.com/shadow1ng/fscan/util"
	"github.com/tomatome/grdp/glog"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	glog.SetLevel(glog.INFO)
	glog.SetLogger(log.New(os.Stdout, "", 0))

	util.CreateYamlFactory("config.yml", &config.Config)
	client.InitMinio(config.Config.Minio)
	client.InitRedis(config.Config.Redis)
	client.InitMysql(config.Config.Mysql)

	WebScan.LoadAllPocs()

	if config.Config.Scan.TemplateSyncStrategy == config.INTERVAL {
		syncPocTemplateTicker := time.NewTicker(time.Hour)
		defer syncPocTemplateTicker.Stop()
		for {
			select {
			case <-syncPocTemplateTicker.C:
				glog.Infof("########################################### [定时]准备同步漏洞模板 ###########################################")
				WebScan.SyncDirtyPocs()
			}
		}
	}

	_ = routers.InitApiRouter().Run(":" + strconv.Itoa(config.Config.Port))
}
