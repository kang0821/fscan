package Plugins

import (
	"fmt"
	"github.com/shadow1ng/fscan/WebScan"
	"github.com/shadow1ng/fscan/WebScan/lib"
	"github.com/shadow1ng/fscan/common"
	"github.com/tomatome/grdp/glog"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func Scan(configInfo *common.ConfigInfo, hostInfo common.HostInfo) {
	glog.Info("start infoscan")
	WebScan.SyncDirtyPocs()
	Hosts, err := common.ParseIP(configInfo, &hostInfo)
	if err != nil {
		panic(err)
	}
	err = lib.Inithttp(configInfo)
	if err != nil {
		panic(err)
	}
	var ch = make(chan struct{}, configInfo.Threads)
	var wg = sync.WaitGroup{}
	web := strconv.Itoa(common.PORTList["web"])
	ms17010 := strconv.Itoa(common.PORTList["ms17010"])
	if len(Hosts) > 0 || len(configInfo.HostPort) > 0 {
		if configInfo.NoPing == false && len(Hosts) > 1 || configInfo.Scantype == "icmp" {
			Hosts = CheckLive(configInfo, Hosts)
			fmt.Println("[*] Icmp alive hosts len is:", len(Hosts))
		}
		if configInfo.Scantype == "icmp" {
			configInfo.LogInfo.LogWG.Wait()
			return
		}
		var AlivePorts []string
		if configInfo.Scantype == "webonly" || configInfo.Scantype == "webpoc" {
			AlivePorts = NoPortScan(Hosts, configInfo.WebPorts, configInfo.NoPorts)
		} else if configInfo.Scantype == "hostname" {
			configInfo.WebPorts = "139"
			AlivePorts = NoPortScan(Hosts, configInfo.WebPorts, configInfo.NoPorts)
		} else if len(Hosts) > 0 {
			AlivePorts = PortScan(Hosts, configInfo)
			fmt.Println("[*] alive ports len is:", len(AlivePorts))
			if configInfo.Scantype == "portscan" {
				configInfo.LogInfo.LogWG.Wait()
				return
			}
		}
		if len(configInfo.HostPort) > 0 {
			AlivePorts = append(AlivePorts, configInfo.HostPort...)
			AlivePorts = common.RemoveDuplicate(AlivePorts)
			configInfo.HostPort = nil
			fmt.Println("[*] AlivePorts len is:", len(AlivePorts))
		}
		var severports []string //severports := []string{"21","22","135"."445","1433","3306","5432","6379","9200","11211","27017"...}
		for _, port := range common.PORTList {
			severports = append(severports, strconv.Itoa(port))
		}
		fmt.Println("start vulscan")
		for _, targetIP := range AlivePorts {
			hostInfo.Host, hostInfo.Ports = strings.Split(targetIP, ":")[0], strings.Split(targetIP, ":")[1]
			if configInfo.Scantype == "all" || configInfo.Scantype == "main" {
				switch {
				case hostInfo.Ports == "135":
					AddScan(hostInfo.Ports, configInfo, hostInfo, &ch, &wg) //findnet
					if configInfo.IsWmi {
						AddScan("1000005", configInfo, hostInfo, &ch, &wg) //wmiexec
					}
				case hostInfo.Ports == "445":
					AddScan(ms17010, configInfo, hostInfo, &ch, &wg) //ms17010
					//AddScan(configInfo.WebPorts, configInfo, ch, &wg)  //smb
					//AddScan("1000002", configInfo, ch, &wg) //smbghost
				case hostInfo.Ports == "9000":
					AddScan(web, configInfo, hostInfo, &ch, &wg)            //http
					AddScan(hostInfo.Ports, configInfo, hostInfo, &ch, &wg) //fcgiscan
				case IsContain(severports, hostInfo.Ports):
					AddScan(hostInfo.Ports, configInfo, hostInfo, &ch, &wg) //plugins scan
				default:
					AddScan(web, configInfo, hostInfo, &ch, &wg) //webtitle
				}
			} else {
				scantype := strconv.Itoa(common.PORTList[configInfo.Scantype])
				AddScan(scantype, configInfo, hostInfo, &ch, &wg)
			}
		}
	}
	for _, url := range configInfo.Urls {
		hostInfo.Url = url
		AddScan(web, configInfo, hostInfo, &ch, &wg)
	}
	wg.Wait()
	configInfo.LogInfo.LogWG.Wait()
	close(configInfo.LogInfo.Results)
	fmt.Printf("已完成 %v/%v\n", configInfo.LogInfo.End, configInfo.LogInfo.Num)
}

var Mutex = &sync.Mutex{}

func AddScan(scantype string, configInfo *common.ConfigInfo, hostInfo common.HostInfo, ch *chan struct{}, wg *sync.WaitGroup) {
	*ch <- struct{}{}
	wg.Add(1)
	go func() {
		Mutex.Lock()
		configInfo.LogInfo.Num += 1
		Mutex.Unlock()
		ScanFunc(&scantype, configInfo, &hostInfo)
		Mutex.Lock()
		configInfo.LogInfo.End += 1
		Mutex.Unlock()
		wg.Done()
		<-*ch
	}()
}

func ScanFunc(name *string, configInfo *common.ConfigInfo, hostInfo *common.HostInfo) {
	f := reflect.ValueOf(PluginList[*name])
	in := []reflect.Value{reflect.ValueOf(configInfo), reflect.ValueOf(hostInfo)}
	f.Call(in)
}

func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}
