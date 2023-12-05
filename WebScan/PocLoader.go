package WebScan

import (
	"github.com/shadow1ng/fscan/WebScan/lib"
	"github.com/shadow1ng/fscan/common"
	"github.com/shadow1ng/fscan/config"
	"github.com/shadow1ng/fscan/model/entity"
	"github.com/tomatome/grdp/glog"
	"gopkg.in/yaml.v2"
)

var AllPocs = make(map[string]*lib.Poc)

// LoadAllPocs	预加载所有漏洞
func LoadAllPocs() {
	var flawList []entity.FwFlaw
	glog.Info("########################################### 准备预加载所有漏洞模板... ###########################################")
	//client.Context.Mysql.MysqlClient.Find(&flawList, entity.FwFlaw{DELETED: false})
	flawList = append(flawList, entity.FwFlaw{
		CODE:   "123",
		NAME:   "123",
		CONFIG: "name: poc-yaml-backup-file\nset:\n  host: request.url.domain\nsets:\n  path:\n    - \"sql\"\n    - \"www\"\n    - \"wwwroot\"\n    - \"index\"\n    - \"backup\"\n    - \"back\"\n    - \"data\"\n    - \"web\"\n    - \"db\"\n    - \"database\"\n    - \"ftp\"\n    - \"admin\"\n    - \"upload\"\n    - \"package\"\n    - \"sql\"\n    - \"old\"\n    - \"test\"\n    - \"root\"\n    - \"beifen\"\n    - host\n  ext:\n    - \"zip\"\n    - \"7z\"\n    - \"rar\"\n    - \"gz\"\n    - \"tar.gz\"\n    - \"db\"\n    - \"bak\"\n\nrules:\n  - method: GET\n    path: /{{path}}.{{ext}}\n    follow_redirects: false\n    continue: true\n    expression: |\n      response.content_type.contains(\"application/\") &&\n      (response.body.startsWith(\"377ABCAF271C\".hexdecode()) ||\n      response.body.startsWith(\"314159265359\".hexdecode()) ||\n      response.body.startsWith(\"53514c69746520666f726d6174203300\".hexdecode()) ||\n      response.body.startsWith(\"1f8b\".hexdecode()) ||\n      response.body.startsWith(\"526172211A0700\".hexdecode()) ||\n      response.body.startsWith(\"FD377A585A0000\".hexdecode()) ||\n      response.body.startsWith(\"1F9D\".hexdecode()) ||\n      response.body.startsWith(\"1FA0\".hexdecode()) ||\n      response.body.startsWith(\"4C5A4950\".hexdecode()) ||\n      response.body.startsWith(\"504B0304\".hexdecode()) )\n#      - \"377ABCAF271C\"  # 7z\n#      - \"314159265359\"  # bz2\n#      - \"53514c69746520666f726d6174203300\"  # SQLite format 3.\n#      - \"1f8b\"  # gz tar.gz\n#      - \"526172211A0700\"  # rar RAR archive version 1.50\n#      - \"526172211A070100\"  # rar RAR archive version 5.0\n#      - \"FD377A585A0000\"  # xz tar.xz\n#      - \"1F9D\"  # z tar.z\n#      - \"1FA0\"  # z tar.z\n#      - \"4C5A4950\"  # lz\n#      - \"504B0304\"  # zip\ndetail:\n  author: shadown1ng(https://github.com/shadown1ng)",
	})
	for _, flaw := range flawList {
		loadPoc(flaw)
	}
	glog.Infof("########################################### 共加载了[%d]项漏洞 ###########################################", len(AllPocs))
}

/*
SyncDirtyPocs 同步发生变更的漏洞。	该方法目的：保证每次漏扫时，使用到的漏洞模板一定是数据库最新的。	TODO 加互斥锁
思路：

	业务系统的会维护一个发生变更的漏洞记录到redis中的SET里（分别有两组：一组是变更的、一组是删除的），键为漏扫引擎服务器节点的IP，值为发生变更的漏洞编号。
	当有漏洞增/删/改时，都会将该漏洞编号放到该SET结构里，每次调用漏洞扫描时，都会先从redis中获取该SET，
	判断当前服务节点是否有需要更新的模板，如果SET结构里有值，则先调用该方法同步发生变更的漏洞模板到该节点内存（替换/新增或删除）。

redis缓存结构（nodeIp: []pocCode）：

	node1: [poc1,poc3,poc105,..]
	node2: []
	node3: [poc3,poc4]
*/
func SyncDirtyPocs() {
	if config.Config.Scan.TemplateSyncStrategy != config.ALWAYS {
		return
	}
	nodeIp := common.GetIP()
	dirtyCacheKey := "imp-service::flaw::dirty::merge" + nodeIp
	removeCacheKey := "imp-service::flaw::dirty::remove" + nodeIp
	mergePocs, err := common.Context.Redis.HGetAll(dirtyCacheKey) // 发生变更的漏洞
	if err != nil {
		glog.Error("获取变更的漏洞信息时出错：%s", err.Error())
		return
	}

	// 去mysql漏洞库获取到所有变更的漏洞
	if len(mergePocs) > 0 {
		var dirtyFlaws []entity.FwFlaw
		common.Context.Mysql.MysqlClient.Where("CODE IN (?) AND DELETED = 0", mergePocs).Find(&dirtyFlaws)
		if len(dirtyFlaws) > 0 {
			var flawNames []string
			for _, mergePoc := range dirtyFlaws {
				flawNames = append(flawNames, mergePoc.NAME)
				loadPoc(mergePoc)
			}
			glog.Infof("########################################### 本次检测到[%d]项发生变更的漏洞。分别是：%s ###########################################", len(dirtyFlaws), flawNames)
		}
		// 加载成功后，清空redis缓存记录
		common.Context.Redis.Del(dirtyCacheKey)
	}

	removePocs, err := common.Context.Redis.HGetAll(removeCacheKey) // 被删除的漏洞
	if err != nil {
		glog.Error("获取移除的漏洞信息时出错：%s", err.Error())
		return
	}
	// 删除的
	if len(removePocs) > 0 {
		for _, removePoc := range removePocs {
			delete(AllPocs, removePoc)
		}
		glog.Infof("########################################### 本次检测到[%d]项被删除的漏洞。分别是：%s ###########################################", len(removePocs), removePocs)
		// 删除成功后，清空redis缓存记录
		common.Context.Redis.Del(removeCacheKey)
	}
}

func loadPoc(flaw entity.FwFlaw) {
	if flaw.CONFIG == "" {
		glog.Infof("漏洞模板[%s]内容为空！", flaw.NAME)
		return
	}
	poc := &lib.Poc{}
	err := yaml.Unmarshal([]byte(flaw.CONFIG), poc)
	if err != nil {
		glog.Infof("解析漏洞模板[%s]出错：%s", flaw.NAME, err.Error())
		return
	}
	poc.Code = flaw.CODE
	AllPocs[poc.Code] = poc
}
