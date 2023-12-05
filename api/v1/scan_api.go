package v1

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/shadow1ng/fscan/Plugins"
	"github.com/shadow1ng/fscan/common"
	"github.com/shadow1ng/fscan/model/request"
	"github.com/shadow1ng/fscan/model/response"
	"github.com/tomatome/grdp/glog"
	"io"
	"net/http"
	"runtime/debug"
	"time"
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
		common.ScanTaskHolder[scanRequest.ConfigInfo.RecordId] = &common.ScanTask{
			TaskId:    scanRequest.ConfigInfo.TaskId,
			RecordId:  scanRequest.ConfigInfo.RecordId,
			Status:    common.SCANNING,
			StartTime: time.Now().Unix(),
		}
		var fileUrl string
		defer func() {
			common.ScanTaskHolder[scanRequest.ConfigInfo.RecordId].EndTime = time.Now().Unix()
			common.ScanTaskHolder[scanRequest.ConfigInfo.RecordId].Status = common.DONE
			if err := recover(); err != nil {
				glog.Errorf("漏洞扫描出现异常: %v\n", err)
				debug.PrintStack()
			}
			// 推送结果
			sendNotify(&scanRequest.ConfigInfo, string(common.DONE), fileUrl)
		}()
		Plugins.Scan(&scanRequest.ConfigInfo, scanRequest.HostInfo)
		// 上传扫描结果文件到minio
		fileUrl, _ = common.Context.Minio.Upload(scanRequest.ConfigInfo.Outputfile)
	}()
	response.Ok(c)
}

func sendNotify(configInfo *common.ConfigInfo, scanProgress, fileUrl string) {
	param, _ := json.Marshal(map[string]string{
		"taskId":       configInfo.TaskId,
		"recordId":     configInfo.RecordId,
		"scanProgress": scanProgress,
		"fileUrl":      fileUrl,
	})
	resp, err := http.Post("POST", configInfo.NotifyUrl, bytes.NewReader(param))
	if err != nil {
		glog.Errorf("推送扫描进度失败: %v\n", err)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("推送扫描进度结果解析失败: %v\n", err)
		return
	}
	glog.Infof("扫描进度推送结果：%s", body)
}
