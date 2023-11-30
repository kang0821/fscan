package Plugins

import (
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
)

func SmbScan2(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return nil
	}
	hasprint := false
	starttime := time.Now().Unix()
	hash := configInfo.HashBytes
	for _, user := range common.Userdict["smb"] {
	PASS:
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err, flag2 := Smb2Con(configInfo, hostInfo, user, pass, hash, hasprint)
			if flag2 {
				hasprint = true
			}
			if flag == true {
				var result string
				if configInfo.Domain != "" {
					result = fmt.Sprintf("[+] SMB2 %v:%v:%v\\%v ", hostInfo.Host, hostInfo.Ports, configInfo.Domain, user)
				} else {
					result = fmt.Sprintf("[+] SMB2 %v:%v:%v ", hostInfo.Host, hostInfo.Ports, user)
				}
				if len(hash) > 0 {
					result += "hash: " + configInfo.Hash
				} else {
					result += pass
				}
				common.LogSuccess(&configInfo.LogInfo, result)
				return err
			} else {
				var errlog string
				if len(configInfo.Hash) > 0 {
					errlog = fmt.Sprintf("[-] smb2 %v:%v %v %v %v", hostInfo.Host, 445, user, configInfo.Hash, err)
				} else {
					errlog = fmt.Sprintf("[-] smb2 %v:%v %v %v %v", hostInfo.Host, 445, user, pass, err)
				}
				errlog = strings.Replace(errlog, "\n", " ", -1)
				common.LogError(&configInfo.LogInfo, errlog)
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["smb"])*len(common.Passwords)) * configInfo.Timeout) {
					return err
				}
			}
			if len(configInfo.Hash) > 0 {
				break PASS
			}
		}
	}
	return tmperr
}

func Smb2Con(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string, hash []byte, hasprint bool) (flag bool, err error, flag2 bool) {
	conn, err := net.DialTimeout("tcp", hostInfo.Host+":445", time.Duration(configInfo.Timeout)*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()
	initiator := smb2.NTLMInitiator{
		User:   user,
		Domain: configInfo.Domain,
	}
	if len(hash) > 0 {
		initiator.Hash = hash
	} else {
		initiator.Password = pass
	}
	d := &smb2.Dialer{
		Initiator: &initiator,
	}

	s, err := d.Dial(conn)
	if err != nil {
		return
	}
	defer s.Logoff()
	names, err := s.ListSharenames()
	if err != nil {
		return
	}
	if !hasprint {
		var result string
		if configInfo.Domain != "" {
			result = fmt.Sprintf("[*] SMB2-shares %v:%v:%v\\%v ", hostInfo.Host, hostInfo.Ports, configInfo.Domain, user)
		} else {
			result = fmt.Sprintf("[*] SMB2-shares %v:%v:%v ", hostInfo.Host, hostInfo.Ports, user)
		}
		if len(hash) > 0 {
			result += "hash: " + configInfo.Hash
		} else {
			result += pass
		}
		result = fmt.Sprintf("%v shares: %v", result, names)
		common.LogSuccess(&configInfo.LogInfo, result)
		flag2 = true
	}
	fs, err := s.Mount("C$")
	if err != nil {
		return
	}
	defer fs.Umount()
	path := `Windows\win.ini`
	f, err := fs.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	defer f.Close()
	flag = true
	return
	//bs, err := ioutil.ReadAll(f)
	//if err != nil {
	//	return
	//}
	//fmt.Println(string(bs))
	//return

}

//if info.Path == ""{
//}
//path = info.Path
//f, err := fs.OpenFile(path, os.O_RDONLY, 0666)
//if err != nil {
//	return
//}
//flag = true
//_, err = f.Seek(0, io.SeekStart)
//if err != nil {
//	return
//}
//bs, err := ioutil.ReadAll(f)
//if err != nil {
//	return
//}
//fmt.Println(string(bs))
//return
//f, err := fs.Create(`Users\Public\Videos\hello.txt`)
//if err != nil {
//	return
//}
//flag = true
//
//_, err = f.Write([]byte("Hello world!"))
//if err != nil {
//	return
//}
//
//_, err = f.Seek(0, io.SeekStart)
//if err != nil {
//	return
//}
//bs, err := ioutil.ReadAll(f)
//if err != nil {
//	return
//}
//fmt.Println(string(bs))
//return
