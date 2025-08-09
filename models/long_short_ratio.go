package models

import (
	"time"

	"gorm.io/gorm"
)

// LongShortRatio 多空比数据模型
type LongShortRatio struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Exchange  string    `json:"exchange" gorm:"index;not null"`  // 交易所名称 (binance, okx)
	Symbol    string    `json:"symbol" gorm:"index;not null"`    // 交易对 (BTC, ETH)
	Ratio     float64   `json:"ratio" gorm:"not null"`           // 多空比值
	Timestamp time.Time `json:"timestamp" gorm:"index;not null"` // 数据时间戳
}

// LongShortRatioRepository 多空比数据仓库
type LongShortRatioRepository struct {
	db *gorm.DB
}

// NewLongShortRatioRepository 创建新的多空比数据仓库
func NewLongShortRatioRepository(db *gorm.DB) *LongShortRatioRepository {
	return &LongShortRatioRepository{db: db}
}

// Create 创建新的多空比记录
func (r *LongShortRatioRepository) Create(ratio *LongShortRatio) error {
	return r.db.Create(ratio).Error
}

// CreateOrUpdate 创建或更新多空比记录（避免重复）
func (r *LongShortRatioRepository) CreateOrUpdate(ratio *LongShortRatio) error {
	// 检查是否已存在相同的记录
	var existing LongShortRatio
	err := r.db.Where("exchange = ? AND symbol = ? AND timestamp = ?",
		ratio.Exchange, ratio.Symbol, ratio.Timestamp).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 记录不存在，创建新记录
		return r.db.Create(ratio).Error
	} else if err != nil {
		// 其他错误
		return err
	} else {
		// 记录已存在，更新比值
		return r.db.Model(&existing).Update("ratio", ratio.Ratio).Error
	}
}

// GetByExchangeAndSymbol 根据交易所和交易对获取最近的多空比数据
func (r *LongShortRatioRepository) GetByExchangeAndSymbol(exchange, symbol string, limit int) ([]LongShortRatio, error) {
	var ratios []LongShortRatio
	err := r.db.Where("exchange = ? AND symbol = ?", exchange, symbol).
		Order("timestamp DESC").
		Limit(limit).
		Find(&ratios).Error
	return ratios, err
}

// GetRecentData 获取最近指定时间范围内的数据
func (r *LongShortRatioRepository) GetRecentData(exchange, symbol string, since time.Time) ([]LongShortRatio, error) {
	var ratios []LongShortRatio
	err := r.db.Where("exchange = ? AND symbol = ? AND timestamp >= ?", exchange, symbol, since).
		Order("timestamp ASC").
		Find(&ratios).Error
	return ratios, err
}

// GetLatest 获取最新的多空比数据
func (r *LongShortRatioRepository) GetLatest(exchange, symbol string) (*LongShortRatio, error) {
	var ratio LongShortRatio
	err := r.db.Where("exchange = ? AND symbol = ?", exchange, symbol).
		Order("timestamp DESC").
		First(&ratio).Error
	if err != nil {
		return nil, err
	}
	return &ratio, nil
}

// DeleteOldData 删除指定时间之前的旧数据
func (r *LongShortRatioRepository) DeleteOldData(before time.Time) error {
	return r.db.Where("timestamp < ?", before).Delete(&LongShortRatio{}).Error
}
