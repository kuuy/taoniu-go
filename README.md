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

# 网格

|交易对    |阶段 |目标位    |止损点     |获利点  |出售百分比 |获利百分比 |
|---------|----|---------|-------   |-------|----------|------   |
|AAVEBUSD |1   |169.954  |83.9      |126.92 |1         |0.05     |
|ANKRBUSD |1   |0.0281   |0.02951   |355.9  |1         |0.05     |
|ANKRBUSD |1   |0.02847  |0.0299    |351.3  |1         |0.05     |
|AVAXBUSD |1   |15.84    |16.64     |0.64   |1         |0.05     |
|DOGEBUSD |1   |0.06005  |0.06306   |167    |1         |0.05     | 
|KAVABUSD |1   |1.524    |1.601     |6.6    |1         |0.05     |
|LTCBUSD  |1   |52.82    |55.47     |0.19   |1         |0.05     |

# 网格交易

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

# taoniu-scripts
淘牛脚本
https://github.com/kuuy/taoniu-scripts

相关功能（Features）
|名称            |说明                    |
|--------       |----                    |
|行情实时更新     | websocket live update 24h tickers |

crontab配置
30 0 * * * /root/taoniu-scripts/cryptos/spot/streams.sh

```shell
/root/taoniu-scripts/cryptos/spot/streams.sh
```

# taoniu-config
淘牛配置
https://github.com/kuuy/taoniu-cofnig

相关功能（Features）
|名称         |说明                 |
|--------    |----                 |
|长驻进程守护  | supervisor.d        |
|开机服务     | systemd             |

# taoniu-android
淘牛客户端（android）
https://github.com/kuuy/taoniu-android

相关功能（Features）
|名称               |说明                                             |
|--------          |----                                             | 
|现货交易计划        | binance spot plans                              |
|现货抢先交易        | binance spot tradings scalping                  |
|杠杆网格交易        | binance margin isolated tradings grids          |
|现货抢先交易日报     | binance spot tradings grids daily analysis      |                                               |
|杠杆网格交易日报     | binance margin isolated tradings daily analysis |

# 免责声明
本项目仅为个人交易测试项目，风险意识完全靠自己把握，出现任何交易失误与本项目无关，请谨慎评估交易的合理性。

# USDT捐赠（TRC20）
TTpMqd3SGckaAmuh1v5VuzJEMegRpqkWWQ

# 推荐链接
[币安](https://www.binance.com/en/activity/referral-entry/CPA?fromActivityPage=true&ref=CPA_007BCNAZTA)  
[Vultr](https://www.vultr.com/?ref=9240160)  
[透明代理入门](https://xtls.github.io/document/level-2/transparent_proxy/transparent_proxy.html#%E9%A6%96%E5%85%88-%E6%88%91%E4%BB%AC%E5%85%88%E8%AF%95%E8%AF%95%E5%81%9A%E5%88%B0%E7%AC%AC%E4%B8%80%E9%98%B6%E6%AE%B5)  
[安全解析检查](https://dnssec.vs.uni-due.de/)  
[代理IP](https://github.com/seevik2580/tor-ip-changer)

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://kuuy-super-space-succotash-6v7g5g7wqv34qq6.github.dev/)