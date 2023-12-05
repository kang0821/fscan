package WebScan

import (
	"embed"
	"fmt"
	"github.com/shadow1ng/fscan/WebScan/lib"
	"github.com/shadow1ng/fscan/common"
	"net/http"
	"strings"
)

//go:embed pocs
var Pocs embed.FS

//var once sync.Once

//var AllPocs []*lib.Poc

func WebScan(configInfo *common.ConfigInfo, hostInfo *common.HostInfo) {
	//once.Do(func() {
	//	initpoc(configInfo)
	//})
	buf := strings.Split(hostInfo.Url, "/")
	configInfo.Target = strings.Join(buf[:3], "/")

	if configInfo.PocName != "" {
		Execute(configInfo)
	} else {
		for _, infostr := range hostInfo.Infostr {
			configInfo.PocName = lib.CheckInfoPoc(infostr)
			Execute(configInfo)
		}
	}
}

func Execute(info *common.ConfigInfo) {
	req, err := http.NewRequest("GET", info.Target, nil)
	if err != nil {
		errlog := fmt.Sprintf("[-] webpocinit %v %v", info.Target, err)
		common.LogError(&info.LogInfo, errlog)
		return
	}

	req.Header.Set("User-agent", info.UserAgent)
	req.Header.Set("Accept", info.Accept)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	if info.Cookie != "" {
		req.Header.Set("Cookie", info.Cookie)
	}
	pocs := filterPoc(info.PocName)
	lib.CheckMultiPoc(info, req, pocs)
}

//func initpoc(info *common.ConfigInfo) {
//	if info.PocPath == "" {
//		entries, err := Pocs.ReadDir("pocs")
//		if err != nil {
//			fmt.Printf("[-] init poc error: %v", err)
//			return
//		}
//		for _, one := range entries {
//			path := one.Name()
//			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
//				if poc, _ := lib.LoadPoc(path, Pocs); poc != nil {
//					AllPocs = append(AllPocs, poc)
//				}
//			}
//		}
//	} else {
//		fmt.Println("[+] load poc from " + info.PocPath)
//		err := filepath.Walk(info.PocPath,
//			func(path string, info os.FileInfo, err error) error {
//				if err != nil || info == nil {
//					return err
//				}
//				if !info.IsDir() {
//					if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
//						poc, _ := lib.LoadPocbyPath(path)
//						if poc != nil {
//							AllPocs = append(AllPocs, poc)
//						}
//					}
//				}
//				return nil
//			})
//		if err != nil {
//			fmt.Printf("[-] init poc error: %v", err)
//		}
//	}
//}

func filterPoc(pocname string) (pocs map[string]*lib.Poc) {
	if pocname == "" {
		return AllPocs
	}
	for _, poc := range AllPocs {
		if strings.Contains(poc.Name, pocname) {
			pocs[poc.Code] = poc
		}
	}
	return
}
