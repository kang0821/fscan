package Plugins

import (
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"strings"
	"time"
)

func MemcachedScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (err error) {
	realhost := fmt.Sprintf("%s:%v", hostInfo.Host, hostInfo.Ports)
	client, err := common.WrapperTcpWithTimeout(configInfo.Socks5Proxy, "tcp", realhost, time.Duration(configInfo.Timeout)*time.Second)
	defer func() {
		if client != nil {
			client.Close()
		}
	}()
	if err == nil {
		err = client.SetDeadline(time.Now().Add(time.Duration(configInfo.Timeout) * time.Second))
		if err == nil {
			_, err = client.Write([]byte("stats\n")) //Set the key randomly to prevent the key on the server from being overwritten
			if err == nil {
				rev := make([]byte, 1024)
				n, err := client.Read(rev)
				if err == nil {
					if strings.Contains(string(rev[:n]), "STAT") {
						result := fmt.Sprintf("[+] Memcached %s unauthorized", realhost)
						common.LogSuccess(&configInfo.LogInfo, result)
					}
				} else {
					errlog := fmt.Sprintf("[-] Memcached %v:%v %v", hostInfo.Host, hostInfo.Ports, err)
					common.LogError(&configInfo.LogInfo, errlog)
				}
			}
		}
	}
	return err
}
