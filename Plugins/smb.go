package Plugins

import (
	"errors"
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"github.com/stacktitan/smb/smb"
	"strings"
	"time"
)

func SmbScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return nil
	}
	starttime := time.Now().Unix()
	for _, user := range common.Userdict["smb"] {
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := doWithTimeOut(configInfo, hostInfo, user, pass)
			if flag == true && err == nil {
				var result string
				if configInfo.Domain != "" {
					result = fmt.Sprintf("[+] SMB %v:%v:%v\\%v %v", hostInfo.Host, hostInfo.Ports, configInfo.Domain, user, pass)
				} else {
					result = fmt.Sprintf("[+] SMB %v:%v:%v %v", hostInfo.Host, hostInfo.Ports, user, pass)
				}
				common.LogSuccess(&configInfo.LogInfo, result)
				return err
			} else {
				errlog := fmt.Sprintf("[-] smb %v:%v %v %v %v", hostInfo.Host, 445, user, pass, err)
				errlog = strings.Replace(errlog, "\n", "", -1)
				common.LogError(&configInfo.LogInfo, errlog)
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["smb"])*len(common.Passwords)) * configInfo.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func SmblConn(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string, signal chan struct{}) (flag bool, err error) {
	flag = false
	Host, Username, Password := hostInfo.Host, user, pass
	options := smb.Options{
		Host:        Host,
		Port:        445,
		User:        Username,
		Password:    Password,
		Domain:      configInfo.Domain,
		Workstation: "",
	}

	session, err := smb.NewSession(options, false)
	if err == nil {
		session.Close()
		if session.IsAuthenticated {
			flag = true
		}
	}
	signal <- struct{}{}
	return flag, err
}

func doWithTimeOut(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string) (flag bool, err error) {
	signal := make(chan struct{})
	go func() {
		flag, err = SmblConn(configInfo, hostInfo, user, pass, signal)
	}()
	select {
	case <-signal:
		return flag, err
	case <-time.After(time.Duration(configInfo.Timeout) * time.Second):
		return false, errors.New("time out")
	}
}
