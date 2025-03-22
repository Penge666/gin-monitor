package main

import (
	"github.com/Penge666/gin_monitor/metrics"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// 初始化一个监控器
	m := metrics.GetNewMonitor()
	// 设置监控路径
	m.SetMetricPath("/metrics")
	m.SetExcludePath([]string{"/metrics"})
	//m.SetMetadata(map[string]string{
	//	"app": "my_service_1",
	//})
	// 设置慢查询时间
	m.SetSlowTime(10)
	// 设置请求持续时间的统计阈值
	m.SetReqDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	// 开始拦截
	m.Attach(r)
	r.GET("/product/:id", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("id"),
		})
	})

	_ = r.Run()
}
