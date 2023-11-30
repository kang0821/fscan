package Plugins

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shadow1ng/fscan/WebScan"
	"github.com/shadow1ng/fscan/WebScan/lib"
	"github.com/shadow1ng/fscan/common"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func WebTitle(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) error {
	if configInfo.Scantype == "webpoc" {
		WebScan.WebScan(configInfo, hostInfo)
		return nil
	}
	err, CheckData := GOWebTitle(configInfo, hostInfo)
	hostInfo.Infostr = WebScan.InfoCheck(configInfo, hostInfo, &CheckData)

	if !configInfo.NoPoc && err == nil {
		WebScan.WebScan(configInfo, hostInfo)
	} else {
		errlog := fmt.Sprintf("[-] webtitle %v %v", hostInfo.Url, err)
		common.LogError(&configInfo.LogInfo, errlog)
	}
	return err
}
func GOWebTitle(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) (err error, CheckData []WebScan.CheckDatas) {
	if hostInfo.Url == "" {
		switch hostInfo.Ports {
		case "80":
			hostInfo.Url = fmt.Sprintf("http://%s", hostInfo.Host)
		case "443":
			hostInfo.Url = fmt.Sprintf("https://%s", hostInfo.Host)
		default:
			host := fmt.Sprintf("%s:%s", hostInfo.Host, hostInfo.Ports)
			protocol := GetProtocol(host, configInfo)
			hostInfo.Url = fmt.Sprintf("%s://%s:%s", protocol, hostInfo.Host, hostInfo.Ports)
		}
	} else {
		if !strings.Contains(hostInfo.Url, "://") {
			host := strings.Split(hostInfo.Url, "/")[0]
			protocol := GetProtocol(host, configInfo)
			hostInfo.Url = fmt.Sprintf("%s://%s", protocol, hostInfo.Url)
		}
	}

	err, result, CheckData := geturl(configInfo, hostInfo, 1, CheckData)
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		return
	}

	//有跳转
	if strings.Contains(result, "://") {
		hostInfo.Url = result
		err, result, CheckData = geturl(configInfo, hostInfo, 3, CheckData)
		if err != nil {
			return
		}
	}

	if result == "https" && !strings.HasPrefix(hostInfo.Url, "https://") {
		hostInfo.Url = strings.Replace(hostInfo.Url, "http://", "https://", 1)
		err, result, CheckData = geturl(configInfo, hostInfo, 1, CheckData)
		//有跳转
		if strings.Contains(result, "://") {
			hostInfo.Url = result
			err, _, CheckData = geturl(configInfo, hostInfo, 3, CheckData)
			if err != nil {
				return
			}
		}
	}
	//是否访问图标
	//err, _, CheckData = geturl(configInfo, 2, CheckData)
	if err != nil {
		return
	}
	return
}

func geturl(configInfo *common.ConfigInfo, hostInfo *common.HostInfo, flag int, CheckData []WebScan.CheckDatas) (error, string, []WebScan.CheckDatas) {
	//flag 1 first try
	//flag 2 /favicon.ico
	//flag 3 302
	//flag 4 400 -> https

	Url := hostInfo.Url
	if flag == 2 {
		URL, err := url.Parse(Url)
		if err == nil {
			Url = fmt.Sprintf("%s://%s/favicon.ico", URL.Scheme, URL.Host)
		} else {
			Url += "/favicon.ico"
		}
	}
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return err, "", CheckData
	}
	req.Header.Set("User-agent", configInfo.UserAgent)
	req.Header.Set("Accept", configInfo.Accept)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	if configInfo.Cookie != "" {
		req.Header.Set("Cookie", configInfo.Cookie)
	}
	//if common.Pocinfo.Cookie != "" {
	//	req.Header.Set("Cookie", "rememberMe=1;"+common.Pocinfo.Cookie)
	//} else {
	//	req.Header.Set("Cookie", "rememberMe=1")
	//}
	req.Header.Set("Connection", "close")
	var client *http.Client
	if flag == 1 {
		client = lib.ClientNoRedirect
	} else {
		client = lib.Client
	}

	resp, err := client.Do(req)
	if err != nil {
		return err, "https", CheckData
	}

	defer resp.Body.Close()
	var title string
	body, err := getRespBody(resp)
	if err != nil {
		return err, "https", CheckData
	}
	CheckData = append(CheckData, WebScan.CheckDatas{body, fmt.Sprintf("%s", resp.Header)})
	var reurl string
	if flag != 2 {
		if !utf8.Valid(body) {
			body, _ = simplifiedchinese.GBK.NewDecoder().Bytes(body)
		}
		title = gettitle(body)
		length := resp.Header.Get("Content-Length")
		if length == "" {
			length = fmt.Sprintf("%v", len(body))
		}
		redirURL, err1 := resp.Location()
		if err1 == nil {
			reurl = redirURL.String()
		}
		result := fmt.Sprintf("[*] WebTitle %-25v code:%-3v len:%-6v title:%v", resp.Request.URL, resp.StatusCode, length, title)
		if reurl != "" {
			result += fmt.Sprintf(" 跳转url: %s", reurl)
		}
		common.LogSuccess(&configInfo.LogInfo, result)
	}
	if reurl != "" {
		return nil, reurl, CheckData
	}
	if resp.StatusCode == 400 && !strings.HasPrefix(hostInfo.Url, "https") {
		return nil, "https", CheckData
	}
	return nil, "", CheckData
}

func getRespBody(oResp *http.Response) ([]byte, error) {
	var body []byte
	if oResp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(oResp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		for {
			buf := make([]byte, 1024)
			n, err := gr.Read(buf)
			if err != nil && err != io.EOF {
				return nil, err
			}
			if n == 0 {
				break
			}
			body = append(body, buf...)
		}
	} else {
		raw, err := io.ReadAll(oResp.Body)
		if err != nil {
			return nil, err
		}
		body = raw
	}
	return body, nil
}

func gettitle(body []byte) (title string) {
	re := regexp.MustCompile("(?ims)<title>(.*?)</title>")
	find := re.FindSubmatch(body)
	if len(find) > 1 {
		title = string(find[1])
		title = strings.TrimSpace(title)
		title = strings.Replace(title, "\n", "", -1)
		title = strings.Replace(title, "\r", "", -1)
		title = strings.Replace(title, "&nbsp;", " ", -1)
		if len(title) > 100 {
			title = title[:100]
		}
		if title == "" {
			title = "\"\"" //空格
		}
	} else {
		title = "None" //没有title
	}
	return
}

func GetProtocol(host string, configInfo *common.ConfigInfo) (protocol string) {
	protocol = "http"
	//如果端口是80或443,跳过Protocol判断
	if strings.HasSuffix(host, ":80") || !strings.Contains(host, ":") {
		return
	} else if strings.HasSuffix(host, ":443") {
		protocol = "https"
		return
	}

	socksconn, err := common.WrapperTcpWithTimeout(configInfo.Socks5Proxy, "tcp", host, time.Duration(configInfo.Timeout)*time.Second)
	if err != nil {
		return
	}
	conn := tls.Client(socksconn, &tls.Config{MinVersion: tls.VersionTLS10, InsecureSkipVerify: true})
	defer func() {
		if conn != nil {
			defer func() {
				if err := recover(); err != nil {
					common.LogError(&configInfo.LogInfo, err)
				}
			}()
			conn.Close()
		}
	}()
	conn.SetDeadline(time.Now().Add(time.Duration(configInfo.Timeout) * time.Second))
	err = conn.Handshake()
	if err == nil || strings.Contains(err.Error(), "handshake failure") {
		protocol = "https"
	}
	return protocol
}
