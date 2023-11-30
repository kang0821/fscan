package Plugins

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/shadow1ng/fscan/common"
	"strings"
	"time"
)

func MssqlScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return
	}
	starttime := time.Now().Unix()
	for _, user := range common.Userdict["mssql"] {
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := MssqlConn(configInfo, hostInfo, user, pass)
			if flag == true && err == nil {
				return err
			} else {
				errlog := fmt.Sprintf("[-] mssql %v:%v %v %v %v", hostInfo.Host, hostInfo.Ports, user, pass, err)
				common.LogError(&configInfo.LogInfo, errlog)
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["mssql"])*len(common.Passwords)) * configInfo.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func MssqlConn(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := hostInfo.Host, hostInfo.Ports, user, pass
	dataSourceName := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%v;encrypt=disable;timeout=%v", Host, Username, Password, Port, time.Duration(configInfo.Timeout)*time.Second)
	db, err := sql.Open("mssql", dataSourceName)
	if err == nil {
		db.SetConnMaxLifetime(time.Duration(configInfo.Timeout) * time.Second)
		db.SetConnMaxIdleTime(time.Duration(configInfo.Timeout) * time.Second)
		db.SetMaxIdleConns(0)
		defer db.Close()
		err = db.Ping()
		if err == nil {
			result := fmt.Sprintf("[+] mssql %v:%v:%v %v", Host, Port, Username, Password)
			common.LogSuccess(&configInfo.LogInfo, result)
			flag = true
		}
	}
	return flag, err
}
