package scheduler

import (
	"CurrencyMonitor/database"
	"CurrencyMonitor/models"
	"CurrencyMonitor/services"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// DataScheduler 数据调度器
type DataScheduler struct {
	cron              *cron.Cron
	dataCollectionSvc *services.DataCollectionService
	repo              *models.LongShortRatioRepository
	symbols           []string
}

// NewDataScheduler 创建新的数据调度器
func NewDataScheduler() *DataScheduler {
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	return &DataScheduler{
		cron:              cron.New(),
		dataCollectionSvc: services.NewDataCollectionService(symbols),
		repo:              models.NewLongShortRatioRepository(database.GetDB()),
		symbols:           symbols,
	}
}

// Start 启动调度器
func (s *DataScheduler) Start() error {
	// 每15分钟收集一次数据
	_, err := s.cron.AddFunc("*/15 * * * *", s.collectData)
	if err != nil {
		return fmt.Errorf("添加数据收集任务失败: %w", err)
	}

	// 每天凌晨2点清理7天前的旧数据
	_, err = s.cron.AddFunc("0 2 * * *", s.cleanupOldData)
	if err != nil {
		return fmt.Errorf("添加数据清理任务失败: %w", err)
	}

	s.cron.Start()
	log.Println("数据调度器启动成功")

	// 立即执行一次数据收集
	go s.collectData()

	return nil
}

// Stop 停止调度器
func (s *DataScheduler) Stop() {
	s.cron.Stop()
	log.Println("数据调度器已停止")
}

// collectData 收集数据
func (s *DataScheduler) collectData() {
	log.Println("开始收集多空比数据...")

	// 收集所有交易所的数据
	data, err := s.dataCollectionSvc.CollectAllData()
	if err != nil {
		log.Printf("收集数据失败: %v", err)
		return
	}

	if len(data) == 0 {
		log.Println("没有收集到任何数据")
		return
	}

	// 保存到数据库
	var savedCount int
	var errors []string

	for _, item := range data {
		ratio := &models.LongShortRatio{
			Exchange:  item.Exchange,
			Symbol:    item.Symbol,
			Ratio:     item.Ratio,
			Timestamp: item.Timestamp,
		}

		if err := s.repo.CreateOrUpdate(ratio); err != nil {
			errorMsg := fmt.Sprintf("保存%s-%s数据失败: %v", item.Exchange, item.Symbol, err)
			errors = append(errors, errorMsg)
			log.Println(errorMsg)
		} else {
			savedCount++
		}
	}

	log.Printf("数据收集完成: 收集%d条，保存%d条", len(data), savedCount)

	if len(errors) > 0 {
		log.Printf("保存过程中发生%d个错误", len(errors))
	}
}

// cleanupOldData 清理旧数据
func (s *DataScheduler) cleanupOldData() {
	log.Println("开始清理旧数据...")

	// 删除7天前的数据
	cutoff := time.Now().AddDate(0, 0, -7)
	err := s.repo.DeleteOldData(cutoff)
	if err != nil {
		log.Printf("清理旧数据失败: %v", err)
		return
	}

	log.Printf("旧数据清理完成，删除了%s之前的数据", cutoff.Format("2006-01-02 15:04:05"))
}

// CollectDataNow 立即收集数据（用于手动触发）
func (s *DataScheduler) CollectDataNow() error {
	s.collectData()
	return nil
}

// GetStatus 获取调度器状态
func (s *DataScheduler) GetStatus() map[string]interface{} {
	entries := s.cron.Entries()

	var nextRuns []string
	for _, entry := range entries {
		nextRuns = append(nextRuns, entry.Next.Format("2006-01-02 15:04:05"))
	}

	return map[string]interface{}{
		"running":      len(entries) > 0,
		"tasks_count":  len(entries),
		"next_runs":    nextRuns,
		"symbols":      s.symbols,
		"last_updated": time.Now().Format("2006-01-02 15:04:05"),
	}
}
