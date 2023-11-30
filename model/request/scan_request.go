package request

import "github.com/shadow1ng/fscan/common"

type ScanRequest struct {
	HostInfo   common.HostInfo   `json:"hostInfo" binding:"required"`
	ConfigInfo common.ConfigInfo `json:"configInfo" binding:"required"`
}

func (request *ScanRequest) ResolveRequest() {
	configInfo := &request.ConfigInfo
	configInfo.LogInfo = common.LogInfo{
		Results: make(chan *string),
	}

	common.HandleString(&configInfo.UserAgent, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	common.HandleString(&configInfo.Accept, "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	common.HandleString(&configInfo.WebPorts, common.DefaultPorts)
	common.HandleInt64(&configInfo.Timeout, 3)
	common.HandleString(&configInfo.Scantype, "all")
	common.HandleInt(&configInfo.Threads, 600)
	common.HandleInt(&configInfo.LiveTop, 10)
	common.HandleInt(&configInfo.BruteThread, 1)
	common.HandleInt64(&configInfo.LogInfo.WaitTime, 60)
	common.HandleInt64(&configInfo.WebTimeout, 5)
	common.HandleInt(&configInfo.PocNum, 20)
}
