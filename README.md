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

# 逐仓杠杆（风格交易）

|交易对    |网格ID                |买入价格  |卖出价格   |买入量 |买出量 |状态   |
|---------|---------------------|---------|----------|------|------|------|
|AAVEBUSD |cdb9ckgv5lfbq1lh0krg |84       |88.2      |0.12  |0.12  |已成交 |
|ANKRBUSD |cdb9d0ov5lf47998ar70 |0.0281   |0.02951   |355.9 |355.9 |待出售 |
|ANKRBUSD |cdau82gqr3idnacpht1g |0.02847  |0.0299    |351.3 |351.3 |待成交 |
|AVAXBUSD |cdb9cm0v5lfbpdu88c00 |15.84    |16.64     |0.64  |0.64  |已出售 |
|DOGEBUSD |cdb9d38v5lf46lpj1pp0 |0.06005  |0.06306   |167   |167   |已出售 |
|KAVABUSD |cdblkb0qr3i25oa8pln0 |1.524    |1.601     |6.6   |6.6   |已出售 |
|LTCBUSD  |cdblnq8qr3i20kvelhd0 |52.82    |55.47     |0.19  |0.19  |已出售 |


# Tradingview分析
```sql
select symbol,summary->>'RECOMMENDATION'
from tradingview_cryptos_analysis
where summary->>'RECOMMENDATION' = 'STRONG_BUY'
```

# taoniu-py
淘牛服务端（python）
https://github.com/kuuy/taoniu-py

相关功能（Features）
|名称     |说明                    |
|--------|----                    |
|Tradingview     |刷新每分钟技术相关指标，推荐买卖信号     |
|Binance Spot Tickers     |刷新24hr行情数据           |
|Binance Spot Klines     |刷新当天K线数据              |

# Quick Start
```bash
git clone https://github.com/kuuy/taoniu-go
cd taoniu/cryptos
go run main.go db migrate
go run main.go binance spot klines daily flush
go run main.go binance spot grids open AVAXBUSD 50
go run main.go cron
```
