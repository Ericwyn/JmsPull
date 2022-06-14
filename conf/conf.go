package conf

import (
	"fmt"
	"github.com/Ericwyn/GoTools/date"
	"github.com/Ericwyn/GoTools/file"
	"github.com/Ericwyn/JmsPull/define"
	"github.com/Ericwyn/JmsPull/log"
	"github.com/Ericwyn/JmsPull/vmess"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var jmsMirror = make([]string, 0)

var TypeIp define.SubLinkType = "ip"
var TypeDomain define.SubLinkType = "domain"
var TypeAll define.SubLinkType = "all"

// SubMsg
// 订阅配置内存缓存
var SubMsg = define.SubMsgStruct{}

// ConfigKey
// viper 的 配置 key
var ConfigKey = define.ConfigKey{
	IpSubLink:               "ip-sub-link",
	IpSubLinkUpdateTime:     "ip-sub-link-update-time",
	DomainSubLink:           "domain-sub-link",
	DomainSubLinkUpdateTime: "domain-sub-link-update-time",
	ApiKey:                  "api-key",
}

// viper 的 key
// package 内部访问 key
var configKeyJmsServer = "jms-server"
var configKeyJmsId = "jms-id"

// GetBuffSubLink
// 获取内存/配置文件里面的配置
// 会有延时，但是可以立马返回
func GetBuffSubLink(linkType define.SubLinkType) string {
	if linkType == TypeDomain {
		// 如果错了的话就从配置文件拿历史数据
		if strings.HasPrefix(SubMsg.DomainLink, "错误") {
			return viper.GetString(ConfigKey.DomainSubLink)
		}
		return SubMsg.DomainLink
	} else if linkType == TypeAll {
		domainMsg := GetBuffSubLink(TypeDomain)
		IpMsg := GetBuffSubLink(TypeDomain)

		// 解析出来然后重新构造
		ipDecode := vmess.Base64Decode(IpMsg)
		domainDecode := vmess.Base64Decode(domainMsg)

		return vmess.Base64Encode(ipDecode + "\n" + domainDecode)
	}
	if strings.HasPrefix(SubMsg.IpLink, "错误") {
		return viper.GetString(ConfigKey.IpSubLink)
	}
	return SubMsg.IpLink
}

func GetRealTimeSubLink(linkType define.SubLinkType) string {
	return apiRequestAndSave(linkType == TypeDomain)
}

// RunCornJobOnce
// 执行一次 Corn 订阅任务
func RunCornJobOnce(useDomain bool) {
	apiRequestAndSave(useDomain)
}

// apiRequestAndSave
// 做一次 api 请求，并且保存数据到本地缓存和配置
// 返回获取到的结果
func apiRequestAndSave(useDomain bool) string {
	subMsg, err := requestJmsSubMsg(useDomain)
	timeStr := getTimeNow()
	if err == nil {
		if useDomain {
			SubMsg.DomainLink = subMsg
			SubMsg.DomainLinkUpdateTime = timeStr

			viper.Set(ConfigKey.DomainSubLink, subMsg)
			viper.Set(ConfigKey.DomainSubLinkUpdateTime, timeStr)
		} else {
			SubMsg.IpLink = subMsg
			SubMsg.IpLinkUpdateTime = timeStr

			viper.Set(ConfigKey.IpSubLink, subMsg)
			viper.Set(ConfigKey.IpSubLinkUpdateTime, timeStr)
		}

		SaveConfig()
		log.I("获取最新订阅数据成功")
		log.I("uesDomain:", useDomain, ", link:", subMsg)
		return subMsg
	} else {
		link := "错误:" + err.Error()
		if useDomain {
			SubMsg.DomainLink = link
			SubMsg.DomainLinkUpdateTime = timeStr
		} else {
			SubMsg.IpLink = link
			SubMsg.IpLinkUpdateTime = timeStr
		}
		log.E("获取最新订阅数据失败")
		log.E("uesDomain:", useDomain, ", error:", subMsg)
		return link
	}
}

func requestJmsSubMsg(useDomain bool) (string, error) {
	mirrorArr := getJmsMirror()
	var wg sync.WaitGroup
	var result string
	var err error
	wg.Add(len(mirrorArr))
	for _, mirror := range mirrorArr {
		url := buildJmsRequestUrl(
			mirror,
			useDomain,
			viper.GetString(configKeyJmsServer),
			viper.GetString(configKeyJmsId),
		)
		result, err = httpGet(url, &wg)
		if err == nil {
			return result, nil
		}
	}
	wg.Wait()
	return "", fmt.Errorf("获取订阅结果超时, " + err.Error() + ", " + date.Format(time.Now(), "yyyy-MM-dd HH:mm:ss"))
}

func httpGet(url string, wg *sync.WaitGroup) (string, error) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	return string(s), nil
}

func getJmsMirror() []string {
	if len(jmsMirror) == 0 {
		configFile := file.OpenFile(getExecPath() + "/" + define.JmsMirrorListFileName)
		configFile.ReadLine(func(line string) {
			line = cleanStr(line)
			jmsMirror = append(jmsMirror, line)
		})
	}
	return jmsMirror
}

func getTimeNow() string {
	return date.Format(time.Now(), "yyyy-MM-dd HH:mm:ss")
}

func getExecPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "."
	}

	if strings.Contains(dir, "\\Local\\Temp") || strings.Contains(dir, "/tmp") {
		return "."
	}
	return dir
}

func buildJmsRequestUrl(mirrorUrl string, useDomain bool, server string, id string) string {
	url := mirrorUrl + "/" + define.JmsSubApi + "?service=" + server + "&id=" + id
	url = strings.Replace(url, "https://", "", 1)
	url = strings.Replace(url+"/"+define.JmsSubApi+"?service="+server+"&id="+id,
		"//", "/", -1)
	if useDomain {
		url += "&usedomains=1"
	}
	return "https://" + url
}

func cleanStr(str string) string {
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\r", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, " ", "", -1)
	return str
}

func InitConfig() {
	// 载入配置/默认配置
	loadConfig()

	// 设置本地缓存，先从内存里 load
	SubMsg.IpLink = viper.GetString(ConfigKey.IpSubLink)
	SubMsg.IpLinkUpdateTime = viper.GetString(ConfigKey.IpSubLinkUpdateTime)
	SubMsg.DomainLink = viper.GetString(ConfigKey.DomainSubLink)
	SubMsg.DomainLinkUpdateTime = viper.GetString(ConfigKey.DomainSubLinkUpdateTime)
}

func loadConfig() {
	var initSave = false
	configDir := getExecPath() + "/.conf"
	configName := "config.yaml"
	configFile := file.OpenFile(configDir + "/" + configName)
	if !configFile.Exits() {
		configFile.CreateFile()
		initSave = true
	}

	viper.SetDefault(configKeyJmsServer, "jms-sub-server")
	viper.SetDefault(configKeyJmsId, "jms-sub-id")
	viper.SetDefault(ConfigKey.ApiKey, "JmsPull 接口 key")

	viper.SetDefault(ConfigKey.IpSubLinkUpdateTime, "2022-06-14 17:58:00")
	viper.SetDefault(ConfigKey.IpSubLink, "错误: 尚未初始化")
	viper.SetDefault(ConfigKey.DomainSubLinkUpdateTime, "2022-06-14 17:58:00")
	viper.SetDefault(ConfigKey.DomainSubLink, "错误: 尚未初始化")

	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	err := viper.ReadInConfig()

	if err != nil {
		log.E("载入配置时候出错")
		panic(err)
	}

	if initSave {
		SaveConfig()
		log.I("请在 .conf/config.yaml 里面写入初始配置 jmsid 与 jmsserver")
		os.Exit(-1)
	}

	printConfigs()
}

func SaveConfig() {
	e := viper.WriteConfig()
	if e != nil {
		log.E("配置文件保存失败")
		log.E(e)
	}
}

func printConfigs() {
	configList := []string{
		configKeyJmsServer,
		configKeyJmsId,

		ConfigKey.IpSubLink,
		ConfigKey.IpSubLinkUpdateTime,
		ConfigKey.DomainSubLink,
		ConfigKey.DomainSubLinkUpdateTime,
		ConfigKey.ApiKey,
	}
	for _, key := range configList {
		log.D("config " + key + "  :  " + viper.GetString(key))
	}
}
