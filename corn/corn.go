package corn

import (
	"github.com/Ericwyn/JmsPull/conf"
	"github.com/Ericwyn/JmsPull/log"
	"github.com/go-co-op/gocron"
	"time"
)

var cornFirstFlag = true

func RunCorn(interval int) {
	s := gocron.NewScheduler(time.UTC)

	// 每 28 分钟刷新一次配置
	s.Every(interval).Minutes().Do(func() {
		log.I("开始获取最新订阅数据")
		// 分别获取域名/ip 两种订阅数据
		conf.RunCornJobOnce(true)
		conf.RunCornJobOnce(false)
	})

	s.StartAsync()
}
