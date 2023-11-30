package Plugins

import (
	"errors"
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"github.com/tomatome/grdp/core"
	"github.com/tomatome/grdp/glog"
	"github.com/tomatome/grdp/protocol/nla"
	"github.com/tomatome/grdp/protocol/pdu"
	"github.com/tomatome/grdp/protocol/rfb"
	"github.com/tomatome/grdp/protocol/sec"
	"github.com/tomatome/grdp/protocol/t125"
	"github.com/tomatome/grdp/protocol/tpkt"
	"github.com/tomatome/grdp/protocol/x224"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Brutelist struct {
	user string
	pass string
}

func RdpScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (tmperr error) {
	if configInfo.IsBrute {
		return
	}

	var wg sync.WaitGroup
	var signal bool
	var num = 0
	var all = len(common.Userdict["rdp"]) * len(common.Passwords)
	var mutex sync.Mutex
	brlist := make(chan Brutelist)
	port, _ := strconv.Atoi(hostInfo.Ports)

	for i := 0; i < configInfo.BruteThread; i++ {
		wg.Add(1)
		//go worker(info.Host, info.Domain, port, &wg, brlist, &signal, &num, all, &mutex, info.Timeout)
		go worker(configInfo, hostInfo, port, &wg, brlist, &signal, &num, all, &mutex)
	}

	for _, user := range common.Userdict["rdp"] {
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			brlist <- Brutelist{user, pass}
		}
	}
	close(brlist)
	go func() {
		wg.Wait()
		signal = true
	}()
	for !signal {
	}

	return tmperr
}

func worker(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, port int, wg *sync.WaitGroup, brlist chan Brutelist, signal *bool, num *int, all int, mutex *sync.Mutex) {
	defer wg.Done()
	for one := range brlist {
		if *signal == true {
			return
		}
		go incrNum(num, mutex)
		user, pass := one.user, one.pass
		//flag, err := RdpConn(info.HostFile, info.Domain, user, pass, port, info.Timeout)
		flag, err := RdpConn(configInfo, hostInfo, user, pass, port)
		if flag == true && err == nil {
			var result string
			if configInfo.Domain != "" {
				result = fmt.Sprintf("[+] RDP %v:%v:%v\\%v %v", hostInfo.Host, port, configInfo.Domain, user, pass)
			} else {
				result = fmt.Sprintf("[+] RDP %v:%v:%v %v", hostInfo.Host, port, user, pass)
			}
			common.LogSuccess(&configInfo.LogInfo, result)
			*signal = true
			return
		} else {
			errlog := fmt.Sprintf("[-] (%v/%v) rdp %v:%v %v %v %v", *num, all, hostInfo.Host, port, user, pass, err)
			common.LogError(&configInfo.LogInfo, errlog)
		}
	}
}

func incrNum(num *int, mutex *sync.Mutex) {
	mutex.Lock()
	*num = *num + 1
	mutex.Unlock()
}

func RdpConn(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, user, password string, port int) (bool, error) {
	target := fmt.Sprintf("%s:%d", configInfo.HostFile, port)
	g := NewClient(target, glog.NONE)
	err := g.Login(configInfo, user, password)

	if err == nil {
		return true, nil
	}

	return false, err
}

type Client struct {
	Host string // ip:port
	tpkt *tpkt.TPKT
	x224 *x224.X224
	mcs  *t125.MCSClient
	sec  *sec.Client
	pdu  *pdu.Client
	vnc  *rfb.RFB
}

func NewClient(host string, logLevel glog.LEVEL) *Client {
	glog.SetLevel(logLevel)
	logger := log.New(os.Stdout, "", 0)
	glog.SetLogger(logger)
	return &Client{
		Host: host,
	}
}

func (g *Client) Login(info *common.ConfigInfo, user, pwd string) error {
	conn, err := common.WrapperTcpWithTimeout(info.Socks5Proxy, "tcp", g.Host, time.Duration(info.Timeout)*time.Second)
	if err != nil {
		return fmt.Errorf("[dial err] %v", err)
	}
	defer conn.Close()
	glog.Info(conn.LocalAddr().String())

	g.tpkt = tpkt.New(core.NewSocketLayer(conn), nla.NewNTLMv2(info.Domain, user, pwd))
	g.x224 = x224.New(g.tpkt)
	g.mcs = t125.NewMCSClient(g.x224)
	g.sec = sec.NewClient(g.mcs)
	g.pdu = pdu.NewClient(g.sec)

	g.sec.SetUser(user)
	g.sec.SetPwd(pwd)
	g.sec.SetDomain(info.Domain)
	//g.sec.SetClientAutoReconnect()

	g.tpkt.SetFastPathListener(g.sec)
	g.sec.SetFastPathListener(g.pdu)
	g.pdu.SetFastPathSender(g.tpkt)

	//g.x224.SetRequestedProtocol(x224.PROTOCOL_SSL)
	//g.x224.SetRequestedProtocol(x224.PROTOCOL_RDP)

	err = g.x224.Connect()
	if err != nil {
		return fmt.Errorf("[x224 connect err] %v", err)
	}
	glog.Info("wait connect ok")
	wg := &sync.WaitGroup{}
	breakFlag := false
	wg.Add(1)

	g.pdu.On("error", func(e error) {
		err = e
		glog.Error("error", e)
		g.pdu.Emit("done")
	})
	g.pdu.On("close", func() {
		err = errors.New("close")
		glog.Info("on close")
		g.pdu.Emit("done")
	})
	g.pdu.On("success", func() {
		err = nil
		glog.Info("on success")
		g.pdu.Emit("done")
	})
	g.pdu.On("ready", func() {
		glog.Info("on ready")
		g.pdu.Emit("done")
	})
	g.pdu.On("update", func(rectangles []pdu.BitmapData) {
		glog.Info("on update:", rectangles)
	})
	g.pdu.On("done", func() {
		if breakFlag == false {
			breakFlag = true
			wg.Done()
		}
	})
	wg.Wait()
	return err
}
