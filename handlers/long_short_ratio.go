package handlers

import (
	"CurrencyMonitor/database"
	"CurrencyMonitor/models"
	"CurrencyMonitor/services"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LongShortRatioHandler 多空比处理器
type LongShortRatioHandler struct {
	repo              *models.LongShortRatioRepository
	dataCollectionSvc *services.DataCollectionService
}

// NewLongShortRatioHandler 创建新的多空比处理器
func NewLongShortRatioHandler() *LongShortRatioHandler {
	repo := models.NewLongShortRatioRepository(database.GetDB())
	dataCollectionSvc := services.NewDataCollectionService([]string{"BTCUSDT", "ETHUSDT"})

	return &LongShortRatioHandler{
		repo:              repo,
		dataCollectionSvc: dataCollectionSvc,
	}
}

// GetCurrentRatios 获取当前多空比数据
func (h *LongShortRatioHandler) GetCurrentRatios(c *gin.Context) {
	exchange := c.Query("exchange")
	symbol := c.Query("symbol")

	if exchange == "" || symbol == "" {
		// 获取所有交易所和交易对的最新数据
		exchanges := []string{"binance", "okx"}
		symbols := []string{"BTCUSDT", "ETHUSDT"}

		var results []gin.H
		for _, ex := range exchanges {
			for _, sym := range symbols {
				ratio, err := h.repo.GetLatest(ex, sym)
				if err != nil {
					continue
				}
				results = append(results, gin.H{
					"exchange":  ratio.Exchange,
					"symbol":    ratio.Symbol,
					"ratio":     ratio.Ratio,
					"timestamp": ratio.Timestamp,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    results,
		})
		return
	}

	// 获取指定交易所和交易对的最新数据
	ratio, err := h.repo.GetLatest(exchange, symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "未找到数据",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"exchange":  ratio.Exchange,
			"symbol":    ratio.Symbol,
			"ratio":     ratio.Ratio,
			"timestamp": ratio.Timestamp,
		},
	})
}

// GetHistoricalData 获取历史多空比数据
func (h *LongShortRatioHandler) GetHistoricalData(c *gin.Context) {
	exchange := c.Query("exchange")
	symbol := c.Query("symbol")
	daysStr := c.DefaultQuery("days", "7")

	if exchange == "" || symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "exchange和symbol参数是必需的",
		})
		return
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	since := time.Now().AddDate(0, 0, -days)
	ratios, err := h.repo.GetRecentData(exchange, symbol, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取历史数据失败",
		})
		return
	}

	var results []gin.H
	for _, ratio := range ratios {
		results = append(results, gin.H{
			"ratio":     ratio.Ratio,
			"timestamp": ratio.Timestamp,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// RefreshData 刷新多空比数据
func (h *LongShortRatioHandler) RefreshData(c *gin.Context) {
	// 收集最新数据
	data, err := h.dataCollectionSvc.CollectAllData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "收集数据失败",
		})
		return
	}

	// 保存到数据库
	var savedCount int
	for _, item := range data {
		ratio := &models.LongShortRatio{
			Exchange:  item.Exchange,
			Symbol:    item.Symbol,
			Ratio:     item.Ratio,
			Timestamp: item.Timestamp,
		}

		if err := h.repo.CreateOrUpdate(ratio); err == nil {
			savedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "数据刷新完成",
		"data": gin.H{
			"collected": len(data),
			"saved":     savedCount,
		},
	})
}

// GetDashboardData 获取仪表板数据
func (h *LongShortRatioHandler) GetDashboardData(c *gin.Context) {
	exchanges := []string{"binance", "okx"}
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	var dashboardData []gin.H

	for _, symbol := range symbols {
		symbolData := gin.H{
			"symbol": symbol,
			"data":   []gin.H{},
		}

		for _, exchange := range exchanges {
			// 获取最新数据
			latest, err := h.repo.GetLatest(exchange, symbol)
			if err != nil {
				continue
			}

			// 获取24小时前的数据进行对比
			since24h := time.Now().Add(-24 * time.Hour)
			historical, err := h.repo.GetRecentData(exchange, symbol, since24h)
			if err != nil {
				continue
			}

			var change float64
			if len(historical) > 0 {
				change = latest.Ratio - historical[0].Ratio
			}

			exchangeData := gin.H{
				"exchange":  exchange,
				"ratio":     latest.Ratio,
				"change":    change,
				"timestamp": latest.Timestamp,
			}

			symbolData["data"] = append(symbolData["data"].([]gin.H), exchangeData)
		}

		dashboardData = append(dashboardData, symbolData)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboardData,
	})
}

// GetComparisonData 获取对比数据（同一交易对在不同交易所的数据）
func (h *LongShortRatioHandler) GetComparisonData(c *gin.Context) {
	symbol := c.Query("symbol")
	daysStr := c.DefaultQuery("days", "7")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "symbol参数是必需的",
		})
		return
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	since := time.Now().AddDate(0, 0, -days)
	exchanges := []string{"binance", "okx"}

	var result = gin.H{
		"symbol": symbol,
		"data":   gin.H{},
	}

	for _, exchange := range exchanges {
		ratios, err := h.repo.GetRecentData(exchange, symbol, since)
		if err != nil {
			continue
		}

		var exchangeData []gin.H
		for _, ratio := range ratios {
			exchangeData = append(exchangeData, gin.H{
				"ratio":     ratio.Ratio,
				"timestamp": ratio.Timestamp,
			})
		}

		result["data"].(gin.H)[exchange] = exchangeData
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetChartData 获取图表数据（支持时间粒度）
func (h *LongShortRatioHandler) GetChartData(c *gin.Context) {
	symbol := c.Query("symbol")
	period := c.DefaultQuery("period", "5m")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "symbol参数是必需的",
		})
		return
	}

	// 固定limit为30个点
	limit := 30

	// 验证时间粒度参数
	validPeriods := map[string]bool{
		"5m":  true,
		"15m": true,
		"30m": true,
		"1h":  true,
		"2h":  true,
		"4h":  true,
		"1d":  true,
	}
	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的时间粒度，支持: 5m, 15m, 30m, 1h, 2h, 4h, 1d",
		})
		return
	}

	// 获取Binance数据
	binanceService := services.NewBinanceService()
	binanceData, err := binanceService.GetLongShortRatioHistory(symbol, period, limit)
	if err != nil {
		fmt.Printf("获取Binance数据失败: %v\n", err)
	}

	// 获取OKX数据
	okxService := services.NewOKXService()
	okxData, err := okxService.GetLongShortRatioHistory(symbol, period, limit)
	if err != nil {
		fmt.Printf("获取OKX数据失败: %v\n", err)
	}

	result := gin.H{
		"symbol":  symbol,
		"period":  period,
		"binance": gin.H{},
		"okx":     gin.H{},
	}

	// 处理Binance数据
	if binanceData != nil {
		var binancePoints []gin.H
		for _, item := range binanceData {
			binancePoints = append(binancePoints, gin.H{
				"ratio":     item.Ratio,
				"timestamp": item.Timestamp,
			})
		}
		result["binance"] = binancePoints
	}

	// 处理OKX数据
	if okxData != nil {
		var okxPoints []gin.H
		for _, item := range okxData {
			okxPoints = append(okxPoints, gin.H{
				"ratio":     item.Ratio,
				"timestamp": item.Timestamp,
			})
		}
		result["okx"] = okxPoints
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
