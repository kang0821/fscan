package lib

import (
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"github.com/tomatome/grdp/glog"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	Client           *http.Client
	ClientNoRedirect *http.Client
	dialTimout       = 5 * time.Second
	keepAlive        = 5 * time.Second
)

func Inithttp(info *common.ConfigInfo) error {
	//common.Proxy = "http://127.0.0.1:8080"
	if info.PocNum == 0 {
		info.PocNum = 20
	}
	if info.WebTimeout == 0 {
		info.WebTimeout = 5
	}
	return InitHttpClient(info, time.Duration(info.WebTimeout)*time.Second)
}

func InitHttpClient(info *common.ConfigInfo, Timeout time.Duration) error {
	type DialContext = func(ctx context.Context, network, addr string) (net.Conn, error)
	dialer := &net.Dialer{
		Timeout:   dialTimout,
		KeepAlive: keepAlive,
	}

	tr := &http.Transport{
		DialContext:         dialer.DialContext,
		MaxConnsPerHost:     5,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: info.PocNum * 2,
		IdleConnTimeout:     keepAlive,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS10, InsecureSkipVerify: true},
		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives:   false,
	}

	if info.Socks5Proxy != "" {
		dialSocksProxy, err := common.Socks5Dailer(info.Socks5Proxy, dialer)
		if err != nil {
			return err
		}
		if contextDialer, ok := dialSocksProxy.(proxy.ContextDialer); ok {
			tr.DialContext = contextDialer.DialContext
		} else {
			return errors.New("Failed type assertion to DialContext")
		}
	} else if info.Proxy != "" {
		if info.Proxy == "1" {
			info.Proxy = "http://127.0.0.1:8080"
		} else if info.Proxy == "2" {
			info.Proxy = "socks5://127.0.0.1:1080"
		} else if !strings.Contains(info.Proxy, "://") {
			info.Proxy = "http://127.0.0.1:" + info.Proxy
		}
		if !strings.HasPrefix(info.Proxy, "socks") && !strings.HasPrefix(info.Proxy, "http") {
			return errors.New("no support this proxy")
		}
		u, err := url.Parse(info.Proxy)
		if err != nil {
			return err
		}
		tr.Proxy = http.ProxyURL(u)
	}

	Client = &http.Client{
		Transport: tr,
		Timeout:   Timeout,
	}
	ClientNoRedirect = &http.Client{
		Transport:     tr,
		Timeout:       Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}
	return nil
}

type Poc struct {
	Code   string
	Name   string  `yaml:"name"`
	Set    StrMap  `yaml:"set"`
	Sets   ListMap `yaml:"sets"`
	Rules  []Rules `yaml:"rules"`
	Groups RuleMap `yaml:"groups"`
	Detail Detail  `yaml:"detail"`
}

type MapSlice = yaml.MapSlice

type StrMap []StrItem
type ListMap []ListItem
type RuleMap []RuleItem

type StrItem struct {
	Key, Value string
}

type ListItem struct {
	Key   string
	Value []string
}

type RuleItem struct {
	Key   string
	Value []Rules
}

func (r *StrMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp yaml.MapSlice
	if err := unmarshal(&tmp); err != nil {
		return err
	}
	for _, one := range tmp {
		key, value := one.Key.(string), one.Value.(string)
		*r = append(*r, StrItem{key, value})
	}
	return nil
}

//func (r *RuleItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
//	var tmp yaml.MapSlice
//	if err := unmarshal(&tmp); err != nil {
//		return err
//	}
//	//for _,one := range tmp{
//	//	key,value := one.Key.(string),one.Value.(string)
//	//	*r = append(*r,StrItem{key,value})
//	//}
//	return nil
//}

func (r *RuleMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp1 yaml.MapSlice
	if err := unmarshal(&tmp1); err != nil {
		return err
	}
	var tmp = make(map[string][]Rules)
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	for _, one := range tmp1 {
		key := one.Key.(string)
		value := tmp[key]
		*r = append(*r, RuleItem{key, value})
	}
	return nil
}

func (r *ListMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp yaml.MapSlice
	if err := unmarshal(&tmp); err != nil {
		return err
	}
	for _, one := range tmp {
		key := one.Key.(string)
		var value []string
		for _, val := range one.Value.([]interface{}) {
			v := fmt.Sprintf("%v", val)
			value = append(value, v)
		}
		*r = append(*r, ListItem{key, value})
	}
	return nil
}

type Rules struct {
	Method          string            `yaml:"method"`
	Path            string            `yaml:"path"`
	Headers         map[string]string `yaml:"headers"`
	Body            string            `yaml:"body"`
	Search          string            `yaml:"search"`
	FollowRedirects bool              `yaml:"follow_redirects"`
	Expression      string            `yaml:"expression"`
	Continue        bool              `yaml:"continue"`
}

type Detail struct {
	Author      string   `yaml:"author"`
	Links       []string `yaml:"links"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version"`
}

//func LoadMultiPoc(Pocs embed.FS, pocname string) []*Poc {
//	var pocs []*Poc
//	for _, f := range SelectPoc(Pocs, pocname) {
//		if p, err := LoadPoc(f, Pocs); err == nil {
//			pocs = append(pocs, p)
//		} else {
//			fmt.Println("[-] load poc ", f, " error:", err)
//		}
//	}
//	return pocs
//}

func LoadPoc(fileName string, Pocs embed.FS) (*Poc, error) {
	p := &Poc{}
	yamlFile, err := Pocs.ReadFile("pocs/" + fileName)

	if err != nil {
		glog.Errorf("[-] load poc %s error1: %v\n", fileName, err)
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, p)
	if err != nil {
		glog.Errorf("[-] load poc %s error2: %v\n", fileName, err)
		return nil, err
	}
	return p, err
}

func SelectPoc(Pocs embed.FS, pocname string) []string {
	entries, err := Pocs.ReadDir("pocs")
	if err != nil {
		glog.Error(err)
	}
	var foundFiles []string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), pocname) {
			foundFiles = append(foundFiles, entry.Name())
		}
	}
	return foundFiles
}

func LoadPocbyPath(fileName string) (*Poc, error) {
	p := &Poc{}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		glog.Errorf("[-] load poc %s error3: %v\n", fileName, err)
		return nil, err
	}
	err = yaml.Unmarshal(data, p)
	if err != nil {
		glog.Errorf("[-] load poc %s error4: %v\n", fileName, err)
		return nil, err
	}
	return p, err
}
