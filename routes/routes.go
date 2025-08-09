package routes

import (
	"CurrencyMonitor/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 多空比处理器
	lsrHandler := handlers.NewLongShortRatioHandler()
	// API日志处理器
	logHandler := handlers.NewAPILogHandler()

	// API路由组
	api := r.Group("/api/v1")
	{
		// 多空比相关API
		longShort := api.Group("/long-short")
		{
			longShort.GET("/current", lsrHandler.GetCurrentRatios)
			longShort.GET("/historical", lsrHandler.GetHistoricalData)
			longShort.GET("/comparison", lsrHandler.GetComparisonData)
			longShort.GET("/chart", lsrHandler.GetChartData)
			longShort.POST("/refresh", lsrHandler.RefreshData)
			longShort.GET("/dashboard", lsrHandler.GetDashboardData)
		}

		// API日志相关API
		logs := api.Group("/logs")
		{
			logs.GET("/recent", logHandler.GetRecentLogs)
			logs.GET("/statistics", logHandler.GetStatistics)
		}
	}

	// 前端页面路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{
			"title": "CurrencyMonitor - 多空比监控",
		})
	})

	r.GET("/dashboard", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{
			"title": "仪表板 - CurrencyMonitor",
		})
	})

	r.GET("/logs", func(c *gin.Context) {
		c.HTML(200, "logs.html", gin.H{
			"title": "API日志 - CurrencyMonitor",
		})
	})

	return r
}
