package main

import (
	"CurrencyMonitor/database"
	"CurrencyMonitor/routes"
	"CurrencyMonitor/scheduler"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 创建并启动数据调度器
	dataScheduler := scheduler.NewDataScheduler()
	if err := dataScheduler.Start(); err != nil {
		log.Fatalf("启动数据调度器失败: %v", err)
	}

	// 设置路由
	r := routes.SetupRoutes()

	// 优雅关闭处理
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("收到关闭信号，正在停止服务...")
		dataScheduler.Stop()
		os.Exit(0)
	}()

	// 启动Web服务器
	log.Println("CurrencyMonitor 启动成功！")
	log.Println("访问地址: http://localhost:8080")
	log.Println("仪表板: http://localhost:8080/dashboard")
	log.Println("API文档: http://localhost:8080/api/v1/long-short/")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动Web服务器失败: %v", err)
	}
}
