package common

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func Parse(configInfo *ConfigInfo, hostInfo *HostInfo) {
	ParseUser(configInfo)
	ParsePass(configInfo)
	ParseInput(configInfo, hostInfo)
	ParseScantype(configInfo)
}

func ParseUser(Info *ConfigInfo) {
	if Info.Username == "" && Info.Userfile == "" {
		return
	}
	var Usernames []string
	if Info.Username != "" {
		Usernames = strings.Split(Info.Username, ",")
	}

	if Info.Userfile != "" {
		users, err := Readfile(Info.Userfile)
		if err == nil {
			for _, user := range users {
				if user != "" {
					Usernames = append(Usernames, user)
				}
			}
		}
	}

	Usernames = RemoveDuplicate(Usernames)
	for name := range Userdict {
		Userdict[name] = Usernames
	}
}

func ParsePass(Info *ConfigInfo) {
	var PwdList []string
	if Info.Password != "" {
		passs := strings.Split(Info.Password, ",")
		for _, pass := range passs {
			if pass != "" {
				PwdList = append(PwdList, pass)
			}
		}
		Passwords = PwdList
	}
	if Info.Passfile != "" {
		passs, err := Readfile(Info.Passfile)
		if err == nil {
			for _, pass := range passs {
				if pass != "" {
					PwdList = append(PwdList, pass)
				}
			}
			Passwords = PwdList
		}
	}
	if Info.URL != "" {
		urls := strings.Split(Info.URL, ",")
		TmpUrls := make(map[string]struct{})
		for _, url := range urls {
			if _, ok := TmpUrls[url]; !ok {
				TmpUrls[url] = struct{}{}
				if url != "" {
					Info.Urls = append(Info.Urls, url)
				}
			}
		}
	}
	if Info.UrlFile != "" {
		urls, err := Readfile(Info.UrlFile)
		if err == nil {
			TmpUrls := make(map[string]struct{})
			for _, url := range urls {
				if _, ok := TmpUrls[url]; !ok {
					TmpUrls[url] = struct{}{}
					if url != "" {
						Info.Urls = append(Info.Urls, url)
					}
				}
			}
		}
	}
	if Info.PortFile != "" {
		ports, err := Readfile(Info.PortFile)
		if err == nil {
			newport := ""
			for _, port := range ports {
				if port != "" {
					newport += port + ","
				}
			}
			Info.WebPorts = newport
		}
	}
}

func Readfile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Open %s error, %v\n", filename, err)
		panic(errors.New("Open " + filename + " error"))
	}
	defer file.Close()
	var content []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			content = append(content, scanner.Text())
		}
	}
	return content, nil
}

func ParseInput(configInfo *ConfigInfo, hostInfo *HostInfo) {
	if hostInfo.Host == "" && configInfo.HostFile == "" && configInfo.URL == "" && configInfo.UrlFile == "" {
		fmt.Println("Host is none")
		//flag.Usage()
		panic(errors.New("host is none"))
	}

	if configInfo.BruteThread <= 0 {
		configInfo.BruteThread = 1
	}

	if configInfo.TmpSave == true {
		IsSave = false
	}

	if configInfo.WebPorts == DefaultPorts {
		configInfo.WebPorts += "," + Webport
	}

	if configInfo.PortAdd != "" {
		if strings.HasSuffix(configInfo.WebPorts, ",") {
			configInfo.WebPorts += configInfo.PortAdd
		} else {
			configInfo.WebPorts += "," + configInfo.PortAdd
		}
	}

	if configInfo.UserAdd != "" {
		user := strings.Split(configInfo.UserAdd, ",")
		for a := range Userdict {
			Userdict[a] = append(Userdict[a], user...)
			Userdict[a] = RemoveDuplicate(Userdict[a])
		}
	}

	if configInfo.PassAdd != "" {
		pass := strings.Split(configInfo.PassAdd, ",")
		Passwords = append(Passwords, pass...)
		Passwords = RemoveDuplicate(Passwords)
	}
	if configInfo.Socks5Proxy != "" && !strings.HasPrefix(configInfo.Socks5Proxy, "socks5://") {
		if !strings.Contains(configInfo.Socks5Proxy, ":") {
			configInfo.Socks5Proxy = "socks5://127.0.0.1" + configInfo.Socks5Proxy
		} else {
			configInfo.Socks5Proxy = "socks5://" + configInfo.Socks5Proxy
		}
	}
	if configInfo.Socks5Proxy != "" {
		fmt.Println("Socks5Proxy:", configInfo.Socks5Proxy)
		_, err := url.Parse(configInfo.Socks5Proxy)
		if err != nil {
			fmt.Println("Socks5Proxy parse error:", err)
			panic(errors.New("Socks5Proxy parse error"))
		}
		configInfo.NoPing = true
	}
	if configInfo.Proxy != "" {
		if configInfo.Proxy == "1" {
			configInfo.Proxy = "http://127.0.0.1:8080"
		} else if configInfo.Proxy == "2" {
			configInfo.Proxy = "socks5://127.0.0.1:1080"
		} else if !strings.Contains(configInfo.Proxy, "://") {
			configInfo.Proxy = "http://127.0.0.1:" + configInfo.Proxy
		}
		fmt.Println("Proxy:", configInfo.Proxy)
		if !strings.HasPrefix(configInfo.Proxy, "socks") && !strings.HasPrefix(configInfo.Proxy, "http") {
			panic(errors.New("no support this proxy"))
		}
		_, err := url.Parse(configInfo.Proxy)
		if err != nil {
			fmt.Println("Proxy parse error:", err)
			panic(errors.New("proxy parse error"))
		}
	}

	if configInfo.Hash != "" && len(configInfo.Hash) != 32 {
		fmt.Println("[-] Hash is error,len(hash) must be 32")
		panic(errors.New("[-] Hash is error,len(hash) must be 32"))
	} else {
		var err error
		configInfo.HashBytes, err = hex.DecodeString(configInfo.Hash)
		if err != nil {
			fmt.Println("[-] Hash is error,hex decode error")
			panic(errors.New("[-] Hash is error,hex decode error"))
		}
	}
}

func ParseScantype(Info *ConfigInfo) {
	_, ok := PORTList[Info.Scantype]
	if !ok {
		showmode()
	}
	if Info.Scantype != "all" && Info.WebPorts == DefaultPorts+","+Webport {
		switch Info.Scantype {
		case "wmiexec":
			Info.WebPorts = "135"
		case "wmiinfo":
			Info.WebPorts = "135"
		case "smbinfo":
			Info.WebPorts = "445"
		case "hostname":
			Info.WebPorts = "135,137,139,445"
		case "smb2":
			Info.WebPorts = "445"
		case "web":
			Info.WebPorts = Webport
		case "webonly":
			Info.WebPorts = Webport
		case "ms17010":
			Info.WebPorts = "445"
		case "cve20200796":
			Info.WebPorts = "445"
		case "portscan":
			Info.WebPorts = DefaultPorts + "," + Webport
		case "main":
			Info.WebPorts = DefaultPorts
		default:
			port, _ := PORTList[Info.Scantype]
			Info.WebPorts = strconv.Itoa(port)
		}
		fmt.Println("-m ", Info.Scantype, " start scan the port:", Info.WebPorts)
	}
}

//func CheckErr(text string, err error, flag bool) {
//	if err != nil {
//		fmt.Println("Parse", text, "error: ", err.Error())
//		if flag {
//			if err != ParseIPErr {
//				fmt.Println(ParseIPErr)
//			}
//			os.Exit(0)
//		}
//	}
//}

func showmode() {
	fmt.Println("The specified scan type does not exist")
	fmt.Println("-m")
	for name := range PORTList {
		fmt.Println("   [" + name + "]")
	}
	panic(errors.New("the specified scan type does not exist"))
}
