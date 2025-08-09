package handlers

import (
	"CurrencyMonitor/database"
	"CurrencyMonitor/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// APILogHandler API日志处理器
type APILogHandler struct {
	repo *models.APILogRepository
}

// NewAPILogHandler 创建新的API日志处理器
func NewAPILogHandler() *APILogHandler {
	repo := models.NewAPILogRepository(database.GetDB())
	return &APILogHandler{
		repo: repo,
	}
}

// GetRecentLogs 获取最近的API日志
func (h *APILogHandler) GetRecentLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	exchange := c.Query("exchange")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	var logs []models.APILog
	if exchange != "" {
		logs, err = h.repo.GetByExchange(exchange, limit)
	} else {
		logs, err = h.repo.GetRecent(limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取日志失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// GetStatistics 获取统计信息
func (h *APILogHandler) GetStatistics(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")

	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	stats, err := h.repo.GetStatistics(since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取统计信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
