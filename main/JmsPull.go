package main

import (
	"github.com/Ericwyn/JmsPull/conf"
	"github.com/Ericwyn/JmsPull/corn"
	"github.com/Ericwyn/JmsPull/define"
	"github.com/Ericwyn/JmsPull/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"strings"
)

func main() {
	conf.InitConfig()

	// 开始定时获取配置
	corn.RunCorn(viper.GetInt(conf.ConfigKey.CornInterval))

	// 开启服务
	startHttpServer()
}

func startHttpServer() {
	r := gin.Default()

	// 直接获取订阅链接
	r.GET("/api/sublink", func(c *gin.Context) {
		if !checkKeyParam(c) {
			return
		}

		realTimeFlag := c.Query("now")

		typeParam := c.Query("type")

		var linkType define.SubLinkType

		if typeParam == "d" || typeParam == "domain" || typeParam == "usedomains" {
			linkType = conf.TypeDomain
		} else if typeParam == "all" {
			linkType = conf.TypeAll
		} else {
			linkType = conf.TypeIp
		}

		if realTimeFlag == "1" {
			c.String(200, conf.GetRealTimeSubLink(linkType))
		} else {
			c.String(200, conf.GetBuffSubLink(linkType))
		}

		c.String(200, conf.GetBuffSubLink(linkType))
		return

	})

	// 查看系统状态
	r.GET("/api/health", func(c *gin.Context) {
		if !checkKeyParam(c) {
			return
		}
		healthMsg := define.SystemHealthMsg{
			// 内存里面的数据
			IpSubLinkStatus:         !strings.HasPrefix(conf.SubMsg.IpLink, "错误"),
			IpSubLinkUpdateTime:     conf.SubMsg.IpLinkUpdateTime,
			DomainSubLinkStatus:     !strings.HasPrefix(conf.SubMsg.DomainLink, "错误"),
			DomainSubLinkUpdateTime: conf.SubMsg.DomainLinkUpdateTime,

			// 配置文件里面的数据
			LocalIpSubLinkStatus:         !strings.HasPrefix(conf.ConfigKey.IpSubLink, "错误"),
			LocalIpSubLinkUpdateTime:     viper.GetString(conf.ConfigKey.IpSubLinkUpdateTime),
			LocalDomainSubLinkStatus:     !strings.HasPrefix(conf.ConfigKey.DomainSubLink, "错误"),
			LocalDomainSubLinkUpdateTime: viper.GetString(conf.ConfigKey.DomainSubLinkUpdateTime),
		}
		c.JSON(200, healthMsg)
	})
	r.Run(":38888")
}

func checkKeyParam(c *gin.Context) bool {
	if c.Query("key") != viper.GetString(conf.ConfigKey.ApiKey) {
		log.I("api key error : " + c.Query("key"))
		c.String(404, "404 page not found")
		return false
	}
	return true
}
