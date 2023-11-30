package Plugins

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/shadow1ng/fscan/common"
	"strings"
	"time"
)

func PostgresScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return
	}
	starttime := time.Now().Unix()
	for _, user := range common.Userdict["postgresql"] {
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", string(user), -1)
			flag, err := PostgresConn(configInfo, hostInfo, user, pass)
			if flag == true && err == nil {
				return err
			} else {
				errlog := fmt.Sprintf("[-] psql %v:%v %v %v %v", hostInfo.Host, hostInfo.Ports, user, pass, err)
				common.LogError(&configInfo.LogInfo, errlog)
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["postgresql"])*len(common.Passwords)) * configInfo.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func PostgresConn(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := hostInfo.Host, hostInfo.Ports, user, pass
	dataSourceName := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", Username, Password, Host, Port, "postgres", "disable")
	db, err := sql.Open("postgres", dataSourceName)
	if err == nil {
		db.SetConnMaxLifetime(time.Duration(configInfo.Timeout) * time.Second)
		defer db.Close()
		err = db.Ping()
		if err == nil {
			result := fmt.Sprintf("[+] Postgres:%v:%v:%v %v", Host, Port, Username, Password)
			common.LogSuccess(&configInfo.LogInfo, result)
			flag = true
		}
	}
	return flag, err
}
