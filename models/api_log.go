package models

import (
	"time"

	"gorm.io/gorm"
)

// APILog API请求日志模型
type APILog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`

	Exchange     string `json:"exchange" gorm:"index;not null"` // 交易所名称 (binance, okx)
	Symbol       string `json:"symbol" gorm:"not null"`         // 交易对
	Period       string `json:"period" gorm:"not null"`         // 时间粒度
	Limit        int    `json:"limit" gorm:"not null"`          // 请求数量
	URL          string `json:"url" gorm:"not null"`            // 请求URL
	StatusCode   int    `json:"status_code" gorm:"not null"`    // 响应状态码
	ResponseTime int64  `json:"response_time" gorm:"not null"`  // 响应时间(毫秒)
	DataCount    int    `json:"data_count" gorm:"not null"`     // 返回数据条数
	ErrorMsg     string `json:"error_msg"`                      // 错误信息
	Success      bool   `json:"success" gorm:"not null"`        // 是否成功
}

// APILogRepository API日志数据仓库
type APILogRepository struct {
	db *gorm.DB
}

// NewAPILogRepository 创建新的API日志数据仓库
func NewAPILogRepository(db *gorm.DB) *APILogRepository {
	return &APILogRepository{db: db}
}

// Create 创建新的API日志记录
func (r *APILogRepository) Create(log *APILog) error {
	return r.db.Create(log).Error
}

// GetRecent 获取最近的日志记录
func (r *APILogRepository) GetRecent(limit int) ([]APILog, error) {
	var logs []APILog
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetByExchange 根据交易所获取日志
func (r *APILogRepository) GetByExchange(exchange string, limit int) ([]APILog, error) {
	var logs []APILog
	err := r.db.Where("exchange = ?", exchange).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetStatistics 获取统计信息
func (r *APILogRepository) GetStatistics(since time.Time) (map[string]interface{}, error) {
	var totalRequests int64
	var successRequests int64
	var avgResponseTime float64

	// 总请求数
	err := r.db.Model(&APILog{}).Where("created_at >= ?", since).Count(&totalRequests).Error
	if err != nil {
		return nil, err
	}

	// 成功请求数
	err = r.db.Model(&APILog{}).Where("created_at >= ? AND success = ?", since, true).Count(&successRequests).Error
	if err != nil {
		return nil, err
	}

	// 平均响应时间
	err = r.db.Model(&APILog{}).Where("created_at >= ?", since).Select("AVG(response_time)").Scan(&avgResponseTime).Error
	if err != nil {
		return nil, err
	}

	successRate := float64(0)
	if totalRequests > 0 {
		successRate = float64(successRequests) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":    totalRequests,
		"success_requests":  successRequests,
		"success_rate":      successRate,
		"avg_response_time": avgResponseTime,
	}, nil
}

// DeleteOldLogs 删除旧日志
func (r *APILogRepository) DeleteOldLogs(before time.Time) error {
	return r.db.Where("created_at < ?", before).Delete(&APILog{}).Error
}
