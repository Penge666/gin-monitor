package main

import (
	"github.com/gin-gonic/gin"
	metrics "test/metrics"
)

func main() {
	r := gin.Default()
	// 初始化一个监控器
	m := metrics.GetNewMonitor()
	r.GET("/product/:id", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("id"),
		})
	})

	_ = r.Run()
}
