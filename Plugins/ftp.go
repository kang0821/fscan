package Plugins

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/shadow1ng/fscan/common"
	"strings"
	"time"
)

func FtpScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return
	}
	starttime := time.Now().Unix()
	flag, err := FtpConn(configInfo, hostInfo, "anonymous", "")
	if flag && err == nil {
		return err
	} else {
		errlog := fmt.Sprintf("[-] ftp %v:%v %v %v", hostInfo.Host, hostInfo.Ports, "anonymous", err)
		common.LogError(&configInfo.LogInfo, errlog)
		tmperr = err
		if common.CheckErrs(err) {
			return err
		}
	}

	for _, user := range common.Userdict["ftp"] {
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := FtpConn(configInfo, hostInfo, user, pass)
			if flag && err == nil {
				return err
			} else {
				errlog := fmt.Sprintf("[-] ftp %v:%v %v %v %v", hostInfo.Host, hostInfo.Ports, user, pass, err)
				common.LogError(&configInfo.LogInfo, errlog)
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["ftp"])*len(common.Passwords)) * configInfo.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func FtpConn(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := hostInfo.Host, hostInfo.Ports, user, pass
	conn, err := ftp.DialTimeout(fmt.Sprintf("%v:%v", Host, Port), time.Duration(configInfo.Timeout)*time.Second)
	if err == nil {
		err = conn.Login(Username, Password)
		if err == nil {
			flag = true
			result := fmt.Sprintf("[+] ftp %v:%v:%v %v", Host, Port, Username, Password)
			dirs, err := conn.List("")
			//defer conn.Logout()
			if err == nil {
				if len(dirs) > 0 {
					for i := 0; i < len(dirs); i++ {
						if len(dirs[i].Name) > 50 {
							result += "\n   [->]" + dirs[i].Name[:50]
						} else {
							result += "\n   [->]" + dirs[i].Name
						}
						if i == 5 {
							break
						}
					}
				}
			}
			common.LogSuccess(&configInfo.LogInfo, result)
		}
	}
	return flag, err
}
