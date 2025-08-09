# CurrencyMonitor - 加密货币多空比监控系统

## 项目简介

CurrencyMonitor 是一个面向个人与小团队的加密货币监控与轻量交易助手，专注于提供多空比趋势查看、价格监控和智能提醒功能。

## 功能特性

### ✅ 已实现功能（MVP第1周 - 专业版）

- **多交易所数据支持**: 支持 Binance 和 OKX 两大交易所
- **实时多空比监控**: 每15分钟自动收集BTC和ETH的多空比数据
- **4个独立图表**: BTC/ETH × Binance/OKX 独立展示，固定30个数据点
- **多时间粒度**: 支持 5m, 15m, 30m, 1h, 2h, 4h, 1d
- **实时涨幅计算**: 每个图表显示当前时间粒度的涨跌幅和百分比
- **左侧导航菜单**: 专业的侧边栏导航，支持移动端
- **数据可视化**: 提供美观的Web仪表板，支持历史趋势图表
- **数据持久化**: 使用SQLite数据库存储历史数据
- **自动数据管理**: 定时清理7天前的旧数据
- **API请求日志**: 完整的API调用日志和统计监控页面
- **请求限流**: 避免API限制和429错误
- **RESTful API**: 完整的API接口支持

### 📊 支持的交易对

- BTC/USDT
- ETH/USDT
- 可扩展支持更多交易对

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 编译运行

```bash
go build -o currency_monitor .
./currency_monitor
```

### 3. 访问系统

- **主页面**: http://localhost:8080
- **仪表板**: http://localhost:8080/dashboard
- **API日志**: http://localhost:8080/logs
- **API接口**: http://localhost:8080/api/v1/long-short/

## API 接口

### 获取当前多空比数据
```
GET /api/v1/long-short/current
GET /api/v1/long-short/current?exchange=binance&symbol=BTCUSDT
```

### 获取历史数据
```
GET /api/v1/long-short/historical?exchange=binance&symbol=BTCUSDT&days=7
```

### 刷新数据
```
POST /api/v1/long-short/refresh
```

### 获取仪表板数据
```
GET /api/v1/long-short/dashboard
```

### 获取图表数据（支持时间粒度）
```
GET /api/v1/long-short/chart?symbol=BTCUSDT&period=5m&limit=100
```

### API日志接口
```
GET /api/v1/logs/recent?limit=100&exchange=binance
GET /api/v1/logs/statistics?hours=24
```

## 项目结构

```
CurrencyMonitor/
├── main.go                 # 主程序入口
├── database/               # 数据库相关
│   └── database.go
├── models/                 # 数据模型
│   └── long_short_ratio.go
├── services/               # 业务服务层
│   ├── binance.go          # Binance API服务
│   ├── okx.go              # OKX API服务
│   └── types.go            # 通用类型定义
├── handlers/               # HTTP处理器
│   └── long_short_ratio.go
├── routes/                 # 路由配置
│   └── routes.go
├── scheduler/              # 定时任务
│   └── scheduler.go
└── templates/              # HTML模板
    └── dashboard.html
```

## 技术栈

- **后端**: Go 1.21, Gin Web框架
- **数据库**: SQLite (GORM ORM)
- **定时任务**: robfig/cron
- **前端**: HTML5, CSS3, JavaScript, Chart.js
- **API**: RESTful API设计

## 开发规范

- 遵循 "Effective Go" 编码规范
- 使用 Go modules 进行依赖管理
- 代码注释使用中文
- 错误处理遵循 Go 最佳实践

## 路线图

- [x] 第1周：多空比查看与数据留存、Web界面、定时收集
- [ ] 第2周：4小时突破提醒、事件记录列表
- [ ] 第3周：联合委托（模拟盘）、规则管理与冷却机制
- [ ] 第4周：实盘开关、Telegram通知、使用引导

## 注意事项

1. 本系统仅用于交易辅助，不提供投资建议
2. 数据来源于交易所公开API，可能存在网络延迟
3. 请合理使用API，避免频繁请求被限制

## 贡献指南

欢迎提交Issue和Pull Request来改进项目。

## 许可证

MIT License
