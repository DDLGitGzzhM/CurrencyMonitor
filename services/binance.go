package services

import (
	"CurrencyMonitor/database"
	"CurrencyMonitor/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// BinanceLongShortRatioResponse Binance多空比API响应结构
type BinanceLongShortRatioResponse struct {
	Symbol         string `json:"symbol"`
	LongShortRatio string `json:"longShortRatio"`
	LongAccount    string `json:"longAccount"`
	ShortAccount   string `json:"shortAccount"`
	Timestamp      int64  `json:"timestamp"`
}

// BinanceService Binance服务
type BinanceService struct {
	baseURL string
	client  *http.Client
}

// NewBinanceService 创建新的Binance服务
func NewBinanceService() *BinanceService {
	return &BinanceService{
		baseURL: "https://fapi.binance.com",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLongShortRatio 获取多空比数据
func (b *BinanceService) GetLongShortRatio(symbol string) (*LongShortRatioData, error) {
	return b.GetLongShortRatioWithPeriod(symbol, "5m", 1)
}

// GetLongShortRatioWithPeriod 获取指定周期的多空比数据
func (b *BinanceService) GetLongShortRatioWithPeriod(symbol, period string, limit int) (*LongShortRatioData, error) {
	url := fmt.Sprintf("%s/futures/data/globalLongShortAccountRatio?symbol=%s&period=%s&limit=%d",
		b.baseURL, symbol, period, limit)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求Binance API失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Binance API返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var ratios []BinanceLongShortRatioResponse
	if err := json.Unmarshal(body, &ratios); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if len(ratios) == 0 {
		return nil, fmt.Errorf("没有获取到多空比数据")
	}

	ratio := ratios[0]
	ratioValue, err := strconv.ParseFloat(ratio.LongShortRatio, 64)
	if err != nil {
		return nil, fmt.Errorf("解析多空比数值失败: %w", err)
	}

	return &LongShortRatioData{
		Exchange:  "binance",
		Symbol:    symbol,
		Ratio:     ratioValue,
		Timestamp: time.Unix(ratio.Timestamp/1000, 0),
	}, nil
}

// GetLongShortRatioHistory 获取多空比历史数据
func (b *BinanceService) GetLongShortRatioHistory(symbol, period string, limit int) ([]*LongShortRatioData, error) {
	startTime := time.Now()
	url := fmt.Sprintf("%s/futures/data/globalLongShortAccountRatio?symbol=%s&period=%s&limit=%d",
		b.baseURL, symbol, period, limit)

	// 创建日志记录
	apiLog := &models.APILog{
		Exchange: "binance",
		Symbol:   symbol,
		Period:   period,
		Limit:    limit,
		URL:      url,
	}

	resp, err := b.client.Get(url)
	responseTime := time.Since(startTime).Milliseconds()
	apiLog.ResponseTime = responseTime

	if err != nil {
		apiLog.Success = false
		apiLog.ErrorMsg = err.Error()
		apiLog.StatusCode = 0
		b.saveLog(apiLog)
		return nil, fmt.Errorf("请求Binance API失败: %w", err)
	}
	defer resp.Body.Close()

	apiLog.StatusCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		apiLog.Success = false
		apiLog.ErrorMsg = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
		b.saveLog(apiLog)
		return nil, fmt.Errorf("Binance API返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiLog.Success = false
		apiLog.ErrorMsg = err.Error()
		b.saveLog(apiLog)
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var ratios []BinanceLongShortRatioResponse
	if err := json.Unmarshal(body, &ratios); err != nil {
		apiLog.Success = false
		apiLog.ErrorMsg = err.Error()
		b.saveLog(apiLog)
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	var results []*LongShortRatioData
	for _, ratio := range ratios {
		ratioValue, err := strconv.ParseFloat(ratio.LongShortRatio, 64)
		if err != nil {
			continue
		}

		results = append(results, &LongShortRatioData{
			Exchange:  "binance",
			Symbol:    symbol,
			Ratio:     ratioValue,
			Timestamp: time.Unix(ratio.Timestamp/1000, 0),
		})
	}

	// 记录成功的日志
	apiLog.Success = true
	apiLog.DataCount = len(results)
	b.saveLog(apiLog)

	return results, nil
}

// GetMultipleSymbolsLongShortRatio 获取多个交易对的多空比数据
func (b *BinanceService) GetMultipleSymbolsLongShortRatio(symbols []string) ([]*LongShortRatioData, error) {
	var results []*LongShortRatioData

	for _, symbol := range symbols {
		data, err := b.GetLongShortRatio(symbol)
		if err != nil {
			// 记录错误但继续处理其他交易对
			fmt.Printf("获取%s多空比失败: %v\n", symbol, err)
			continue
		}
		results = append(results, data)

		// 避免请求过于频繁
		time.Sleep(100 * time.Millisecond)
	}

	return results, nil
}

// saveLog 保存API日志
func (b *BinanceService) saveLog(apiLog *models.APILog) {
	if db := database.GetDB(); db != nil {
		repo := models.NewAPILogRepository(db)
		if err := repo.Create(apiLog); err != nil {
			fmt.Printf("保存Binance API日志失败: %v\n", err)
		}
	}
}
