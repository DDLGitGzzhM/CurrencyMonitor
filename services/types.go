package services

import (
	"fmt"
	"time"
)

// LongShortRatioData 多空比数据通用结构
type LongShortRatioData struct {
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"`
	Ratio     float64   `json:"ratio"`
	Timestamp time.Time `json:"timestamp"`
}

// ExchangeService 交易所服务接口
type ExchangeService interface {
	GetLongShortRatio(symbol string) (*LongShortRatioData, error)
	GetMultipleSymbolsLongShortRatio(symbols []string) ([]*LongShortRatioData, error)
}

// DataCollectionService 数据收集服务
type DataCollectionService struct {
	binanceService *BinanceService
	okxService     *OKXService
	symbols        []string
}

// NewDataCollectionService 创建新的数据收集服务
func NewDataCollectionService(symbols []string) *DataCollectionService {
	return &DataCollectionService{
		binanceService: NewBinanceService(),
		okxService:     NewOKXService(),
		symbols:        symbols,
	}
}

// CollectAllData 收集所有交易所的多空比数据
func (d *DataCollectionService) CollectAllData() ([]*LongShortRatioData, error) {
	var allData []*LongShortRatioData

	// 收集Binance数据
	binanceData, err := d.binanceService.GetMultipleSymbolsLongShortRatio(d.symbols)
	if err != nil {
		// 记录错误但继续
		fmt.Printf("收集Binance数据失败: %v\n", err)
	} else {
		allData = append(allData, binanceData...)
	}

	// 收集OKX数据
	okxData, err := d.okxService.GetMultipleSymbolsLongShortRatio(d.symbols)
	if err != nil {
		// 记录错误但继续
		fmt.Printf("收集OKX数据失败: %v\n", err)
	} else {
		allData = append(allData, okxData...)
	}

	return allData, nil
}

// GetDataByExchange 根据交易所获取数据
func (d *DataCollectionService) GetDataByExchange(exchange string) ([]*LongShortRatioData, error) {
	switch exchange {
	case "binance":
		return d.binanceService.GetMultipleSymbolsLongShortRatio(d.symbols)
	case "okx":
		return d.okxService.GetMultipleSymbolsLongShortRatio(d.symbols)
	default:
		return nil, fmt.Errorf("不支持的交易所: %s", exchange)
	}
}
