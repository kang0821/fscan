package Plugins

import (
	"errors"
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

func SshScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return
	}
	starttime := time.Now().Unix()
	for _, user := range common.Userdict["ssh"] {
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := SshConn(configInfo, hostInfo, user, pass)
			if flag == true && err == nil {
				return err
			} else {
				errlog := fmt.Sprintf("[-] ssh %v:%v %v %v %v", hostInfo.Host, hostInfo.Ports, user, pass, err)
				common.LogError(&configInfo.LogInfo, errlog)
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["ssh"])*len(common.Passwords)) * configInfo.Timeout) {
					return err
				}
			}
			if configInfo.SshKey != "" {
				return err
			}
		}
	}
	return tmperr
}

func SshConn(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := hostInfo.Host, hostInfo.Ports, user, pass
	var Auth []ssh.AuthMethod
	if configInfo.SshKey != "" {
		pemBytes, err := ioutil.ReadFile(configInfo.SshKey)
		if err != nil {
			return false, errors.New("read key failed" + err.Error())
		}
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return false, errors.New("parse key failed" + err.Error())
		}
		Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		Auth = []ssh.AuthMethod{ssh.Password(Password)}
	}

	config := &ssh.ClientConfig{
		User:    Username,
		Auth:    Auth,
		Timeout: time.Duration(configInfo.Timeout) * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", Host, Port), config)
	if err == nil {
		defer client.Close()
		session, err := client.NewSession()
		if err == nil {
			defer session.Close()
			flag = true
			var result string
			if configInfo.Command != "" {
				combo, _ := session.CombinedOutput(configInfo.Command)
				result = fmt.Sprintf("[+] SSH %v:%v:%v %v \n %v", Host, Port, Username, Password, string(combo))
				if configInfo.SshKey != "" {
					result = fmt.Sprintf("[+] SSH %v:%v sshkey correct \n %v", Host, Port, string(combo))
				}
				common.LogSuccess(&configInfo.LogInfo, result)
			} else {
				result = fmt.Sprintf("[+] SSH %v:%v:%v %v", Host, Port, Username, Password)
				if configInfo.SshKey != "" {
					result = fmt.Sprintf("[+] SSH %v:%v sshkey correct", Host, Port)
				}
				common.LogSuccess(&configInfo.LogInfo, result)
			}
		}
	}
	return flag, err

}
