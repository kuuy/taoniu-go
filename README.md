# taoniu-go
淘牛服务端（golang）

# 技术指标
|符号     |名称                    |
|--------|----                    |
|ATR     |均幅指标                 |
|PIVOT   |轴点                    |
|KDJ     |随机指标                 |
|BBANDS  |布林带（布林极值、布林带宽）|
|ZLEMA   |零延迟指数平均数          |
|HAZLEMA |平滑零延迟指数平均数       |

# Tradingview分析
```sql
select symbol,summary->>'RECOMMENDATION'
from tradingview_cryptos_analysis
where summary->>'RECOMMENDATION' = 'STRONG_BUY'
```
https://github.com/kuuy/taoniu-py

# Quick Start
```bash
git clone https://github.com/kuuy/taoniu-go
cd taoniu/cryptos
go run main.go db migrate
go run main.go binance spot klines daily flush
go run main.go binance spot grids open AVAXBUSD 50
go run main.go cron
```
