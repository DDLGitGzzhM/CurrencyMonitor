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

// OKXLongShortRatioResponse OKX多空比API响应结构
type OKXLongShortRatioResponse struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"` // OKX返回二维数组格式: [["timestamp", "ratio"], ...]
}

// OKXService OKX服务
type OKXService struct {
	baseURL     string
	client      *http.Client
	lastRequest time.Time
}

// NewOKXService 创建新的OKX服务
func NewOKXService() *OKXService {
	return &OKXService{
		baseURL: "https://www.okx.com",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLongShortRatio 获取多空比数据
func (o *OKXService) GetLongShortRatio(symbol string) (*LongShortRatioData, error) {
	return o.GetLongShortRatioWithPeriod(symbol, "5m", 1)
}

// GetLongShortRatioWithPeriod 获取指定周期的多空比数据
func (o *OKXService) GetLongShortRatioWithPeriod(symbol, period string, limit int) (*LongShortRatioData, error) {
	// 限流：每次请求间隔至少1秒
	if time.Since(o.lastRequest) < time.Second {
		time.Sleep(time.Second - time.Since(o.lastRequest))
	}
	o.lastRequest = time.Now()

	// OKX API使用不同的交易对格式，需要转换
	instId := o.convertSymbol(symbol)
	url := fmt.Sprintf("%s/api/v5/rubik/stat/contracts/long-short-account-ratio?ccy=%s&period=%s&limit=%d",
		o.baseURL, instId, period, limit)

	resp, err := o.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求OKX API失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OKX API返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var response OKXLongShortRatioResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("OKX API返回错误: %s", response.Msg)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("没有获取到多空比数据")
	}

	// OKX返回的数据格式: [["timestamp", "ratio"], ...]
	// 取最新的一条数据（第一个元素）
	data := response.Data[0]
	if len(data) < 2 {
		return nil, fmt.Errorf("OKX数据格式错误")
	}

	ratioValue, err := strconv.ParseFloat(data[1], 64)
	if err != nil {
		return nil, fmt.Errorf("解析多空比数值失败: %w", err)
	}

	timestamp, err := strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("解析时间戳失败: %w", err)
	}

	return &LongShortRatioData{
		Exchange:  "okx",
		Symbol:    symbol,
		Ratio:     ratioValue,
		Timestamp: time.Unix(timestamp/1000, 0),
	}, nil
}

// GetLongShortRatioHistory 获取多空比历史数据
func (o *OKXService) GetLongShortRatioHistory(symbol, period string, limit int) ([]*LongShortRatioData, error) {
	startTime := time.Now()

	// 限流：每次请求间隔至少1秒
	if time.Since(o.lastRequest) < time.Second {
		time.Sleep(time.Second - time.Since(o.lastRequest))
	}
	o.lastRequest = time.Now()

	// OKX API使用不同的交易对格式，需要转换
	instId := o.convertSymbol(symbol)
	// OKX API不支持limit参数，我们获取全部数据然后截取
	url := fmt.Sprintf("%s/api/v5/rubik/stat/contracts/long-short-account-ratio?ccy=%s&period=%s",
		o.baseURL, instId, period)

	// 创建日志记录
	apiLog := &models.APILog{
		Exchange: "okx",
		Symbol:   symbol,
		Period:   period,
		Limit:    limit,
		URL:      url,
	}

	resp, err := o.client.Get(url)
	responseTime := time.Since(startTime).Milliseconds()

	// 更新日志记录
	apiLog.ResponseTime = responseTime

	if err != nil {
		apiLog.Success = false
		apiLog.ErrorMsg = err.Error()
		apiLog.StatusCode = 0
		o.saveLog(apiLog)
		return nil, fmt.Errorf("请求OKX API失败: %w", err)
	}
	defer resp.Body.Close()

	apiLog.StatusCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		apiLog.Success = false
		apiLog.ErrorMsg = fmt.Sprintf("HTTP状态码: %d", resp.StatusCode)
		o.saveLog(apiLog)
		return nil, fmt.Errorf("OKX API返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiLog.Success = false
		apiLog.ErrorMsg = err.Error()
		o.saveLog(apiLog)
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var response OKXLongShortRatioResponse
	if err := json.Unmarshal(body, &response); err != nil {
		apiLog.Success = false
		apiLog.ErrorMsg = err.Error()
		o.saveLog(apiLog)
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if response.Code != "0" {
		apiLog.Success = false
		apiLog.ErrorMsg = response.Msg
		o.saveLog(apiLog)
		return nil, fmt.Errorf("OKX API返回错误: %s", response.Msg)
	}

	var results []*LongShortRatioData
	count := 0
	for _, data := range response.Data {
		if count >= limit {
			break // 限制返回的数据点数量
		}

		if len(data) < 2 {
			continue
		}

		ratioValue, err := strconv.ParseFloat(data[1], 64)
		if err != nil {
			continue
		}

		timestamp, err := strconv.ParseInt(data[0], 10, 64)
		if err != nil {
			continue
		}

		results = append(results, &LongShortRatioData{
			Exchange:  "okx",
			Symbol:    symbol,
			Ratio:     ratioValue,
			Timestamp: time.Unix(timestamp/1000, 0),
		})
		count++
	}

	// 记录成功的日志
	apiLog.Success = true
	apiLog.DataCount = len(results)
	o.saveLog(apiLog)

	return results, nil
}

// GetMultipleSymbolsLongShortRatio 获取多个交易对的多空比数据
func (o *OKXService) GetMultipleSymbolsLongShortRatio(symbols []string) ([]*LongShortRatioData, error) {
	var results []*LongShortRatioData

	for _, symbol := range symbols {
		data, err := o.GetLongShortRatio(symbol)
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

// convertSymbol 转换交易对格式
func (o *OKXService) convertSymbol(symbol string) string {
	// 将BTC转换为BTC，ETH转换为ETH等
	// OKX API使用基础货币名称
	switch symbol {
	case "BTCUSDT":
		return "BTC"
	case "ETHUSDT":
		return "ETH"
	default:
		// 移除USDT后缀
		if len(symbol) > 4 && symbol[len(symbol)-4:] == "USDT" {
			return symbol[:len(symbol)-4]
		}
		return symbol
	}
}

// saveLog 保存API日志
func (o *OKXService) saveLog(apiLog *models.APILog) {
	if db := database.GetDB(); db != nil {
		repo := models.NewAPILogRepository(db)
		if err := repo.Create(apiLog); err != nil {
			fmt.Printf("保存OKX API日志失败: %v\n", err)
		}
	}
}
