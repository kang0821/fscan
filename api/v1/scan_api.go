package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/shadow1ng/fscan/Plugins"
	"github.com/shadow1ng/fscan/common"
	"github.com/shadow1ng/fscan/model/request"
	"github.com/shadow1ng/fscan/model/response"
	"log"
)

type ScanApi struct{}

func (*ScanApi) StartScan(c *gin.Context) {
	var scanRequest request.ScanRequest
	err := c.ShouldBindJSON(&scanRequest)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	/* 参数预处理 */
	scanRequest.ResolveRequest()

	/* 参数解析 */
	common.InitLog(&scanRequest.ConfigInfo)
	common.Parse(&scanRequest.ConfigInfo, &scanRequest.HostInfo)

	/* 将扫描任务放入协程中处理，直接响应。因为扫描过程会非常慢，http调用时会超时 */
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("漏洞扫描出现异常: %v\n", err)
			}
		}()
		Plugins.Scan(&scanRequest.ConfigInfo, scanRequest.HostInfo)
	}()
	response.Ok(c)
}
