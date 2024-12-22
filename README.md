# taoniu-go
淘牛服务端（golang）

互联网从最开始的静态内容消费阶段，蜕变到了动态内容生产传播的阶段，在活跃的互联网活动中，开始涌现出了一批安全的数字资产共识的机制，并衍生出了庞大的加密货币交易市场。  

在低门槛的交易环境中，可以快速地学习和应用很多经典的技术指标，帮助自己更好的理解和掌握市场经营策略。并通过自动交易方式更多、更广地参与市场投资行为。
在对自动交易的监督和调整过程中，不断地优化自己的资金规划意识，让自己的资金能够更安全、有效地运转起来。

淘牛通过Scalping进行最小化交易，最大范围去参与市场交易行为，在深度参与的过程中，尽情享受最宽幅度的市场冲击与回调带来的紧张与兴奋，漫步在整个市场周期的牛市逆转及熊市反转之间。
斐波那契散列的收缩与扩张，让我们更好地去计算潜在交易区间，更合理地进行因果分析。

淘牛提供了非常有效的仓位、对赌计算工具，在应用谐波、威科夫等交易模式中，对任意规模、任意层次的资金进行合理的仓位管理。

淘牛帮你开启慢吸高补，长进短出，筑顶捞底的漫长炒币生活。

# 技术指标
| 符号                |名称                   |
|-------------------|----                   |
| ATR               |均幅指标                 |
| PIVOT             |轴点                    |
| KDJ               |随机指标                 |
| Boll BANDS        |布林带（布林极值、布林带宽） |
| Ichimoku Cloud    |一目均衡图                |
| ZLEMA             |零延迟指数平均数           |
| HA ZLEMA          |平滑零延迟指数平均数        |
| Volume Profile    |成交量分布（控制点，价值区间）|
| Andean Oscillator |安第斯振荡器               |

# Scalping

|交易对    |阶段 |目标位    |止损点     |获利点  |出售百分比 |获利百分比 |
|---------|----|---------|-------   |-------|----------|------   |
|AAVEBUSD |1   |169.954  |83.9      |126.92 |1         |0.05     |
|ANKRBUSD |1   |0.0281   |0.02951   |355.9  |1         |0.05     |
|ANKRBUSD |1   |0.02847  |0.0299    |351.3  |1         |0.05     |
|AVAXBUSD |1   |15.84    |16.64     |0.64   |1         |0.05     |
|DOGEBUSD |1   |0.06005  |0.06306   |167    |1         |0.05     | 
|KAVABUSD |1   |1.524    |1.601     |6.6    |1         |0.05     |
|LTCBUSD  |1   |52.82    |55.47     |0.19   |1         |0.05     |

# Scalping Trading

|交易对    |网格ID                |买入价格  |卖出价格   |买入量 |买出量 |状态   |
|---------|---------------------|---------|----------|------|------|------|
|AAVEBUSD |cdb9ckgv5lfbq1lh0krg |84       |88.2      |0.12  |0.12  |已成交 |
|ANKRBUSD |cdb9d0ov5lf47998ar70 |0.0281   |0.02951   |355.9 |355.9 |待出售 |
|ANKRBUSD |cdau82gqr3idnacpht1g |0.02847  |0.0299    |351.3 |351.3 |待成交 |
|AVAXBUSD |cdb9cm0v5lfbpdu88c00 |15.84    |16.64     |0.64  |0.64  |已出售 |
|DOGEBUSD |cdb9d38v5lf46lpj1pp0 |0.06005  |0.06306   |167   |167   |已出售 |
|KAVABUSD |cdblkb0qr3i25oa8pln0 |1.524    |1.601     |6.6   |6.6   |已出售 |
|LTCBUSD  |cdblnq8qr3i20kvelhd0 |52.82    |55.47     |0.19  |0.19  |已出售 |

# Triggers

|交易对    |价格   |余额      |交易线          |初始额 |目标额 |止损额 | 状态  | 备注          |
|---------|------|---------|----------------|-----|------|------|-------|--------------|
|ZECBUSD  |43.2  |440.078  |[[48.4,47.4 ... |500  |900   |100   |正常    |              |
|AVAXBUSD |17.95 |489.98   |[[18.24,17.97...|500  |900   |100   |正常    |              |
|LTCBUSD  |92.12 |459.98   |[[100.05,98.5...|500  |900   |100   |异常    |APIError ...  |

# Triggers Trading

|交易对    |买入订单   |卖出订单    |买入价格  |卖出价格   |买入量   |卖出量    | 状态  | 备注          |
|---------|----------|--------- |---------|----------|--------|---------|------|--------------|
|STPTBUSD |48477921  |48479037  |0.0426   |0.04275   |234.8   |234      |已完成 |              |
|UNIBUSD  |341576284 |          |6.344    |6.367     |1.58    |1.58     |待出售 |              |
|ZECBUSD  |185880417 |185884874 |44.7     |44.9      |0.224   |0.223    |出售中 |              |
|AVAXBUSD |677600391 |          |17.88    |17.95     |0.56    |0.56     |待出售 |              |
|STPTBUSD |48475525  |48477920  |0.04235  |0.0425    |236.2   |235.3    |已完成 |              |
|ZECBUSD  |185870784 |185875454 |42.8     |43        |0.234   |0.233    |出售中 |              |

# Triggers Trading Analysis

|日期        |买入数   |卖出数    |买入量     |卖出量     |利益       |数据                         |
|-----------|--------|---------|----------|----------|-----------|----------------------------|
|2023-02-10 |280     |200      |2451.87   |2001.38   |1.22       |当天、历史待出售，盈余虚似币量   |

# Tradingview分析
```sql
select symbol,summary->>'RECOMMENDATION'
from tradingview_cryptos_analysis
where summary->>'RECOMMENDATION' = 'STRONG_BUY'
```

# Quick Start
```bash
DESKTOP OS: Ubuntu 22.04.4 LTS
SERVER OS: Debian GNU/Linux 12

sudo apt install postgresql-15 pwgen

sudo su postgres
psql

pwgen -s 16

CREATE DATABASE taoniu;
CREATE USER taoniu WITH ENCRYPTED PASSWORD 'xxxxxxxxxxxxxxxx';
GRANT ALL PRIVILEGES ON DATABASE taoniu TO taoniu;
ALTER DATABASE taoniu OWNER TO taoniu;

mkdir ~/build && cd ~/build
curl https://download.redis.io/releases/redis-6.2.14.tar.gz -o redis-6.2.14.tar.gz
tar xfz redis-6.2.14.tar.gz
cd redis-6.2.14 && make -j20 && make install

cd ~/build
curl https://go.dev/dl/go1.21.13.linux-amd64.tar.gz -o go1.21.13.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz

git clone https://github.com/kuuy/taoniu-go ~/

cd ~/taoniu/cryptos

go build -ldflags "-s -w" -o cryptos

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

cd ~/taoniu-go/cryptos/grpc/protos
protoc --go_out=../ --go_opt=paths=source_relative \
    --go-grpc_out=../ --go-grpc_opt=paths=source_relative \
    **/*.proto

~/taoniu-go/cryptos db migrate
~/taoniu-go/cryptos binance spot klines flush 1d 100
~/taoniu-go/cryptos binance spot positions calc BNBUSDT 400000000 10 519.6 173.241
~/taoniu-go/cryptos binance spot gambling calc BNBUSDT 1 519.6 173.241
~/taoniu-go/cryptos binance futures positions calc BNBUSDT 400000000 10 1 519.6 173.241
~/taoniu-go/cryptos binance futures gambling calc BNBUSDT 1 519.6 173.241
~/taoniu-go/cryptos cron
~/taoniu-go/cryptos api
~/taoniu-go/cryptos grpc
```

# crontab

```
# m h  dom mon dow   command
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance spot tasks account flush ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance spot tasks tradings scalping place ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance spot tasks tradings scalping flush ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance spot tasks tradings triggers place ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance spot tasks tradings triggers flush ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance futures tasks tradings scalping place ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance futures tasks tradings scalping flush ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance futures tasks tradings triggers place ; sleep 5 ; done
* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance futures tasks tradings triggers flush ; sleep 5 ; done

* * * * * /root/taoniu-go/cryptos binance spot tasks symbols flush
*/5 * * * * /root/taoniu-go/cryptos binance spot tasks analysis tradings scalping flush
*/5 * * * * /root/taoniu-go/cryptos binance spot tasks analysis tradings triggers flush
3,8,33,48 * * * * /root/taoniu-go/cryptos binance spot tasks klines clean
7,22,37,52 * * * * /root/taoniu-go/cryptos binance spot tasks strategies clean
11,26,41,56 * * * * /root/taoniu-go/cryptos binance spot tasks plans clean

* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance futures tasks account flush ; sleep 5 ; don

* * * * * /root/taoniu-go/cryptos binance futures tasks symbols flush
*/5 * * * * /root/taoniu-go/cryptos binance futures tasks analysis tradings scalping flush
*/5 * * * * /root/taoniu-go/cryptos binance futures tasks analysis tradings triggers flush
3,8,33,48 * * * * /root/taoniu-go/cryptos binance futures tasks klines clean
7,22,37,52 * * * * /root/taoniu-go/cryptos binance futures tasks strategies clean
11,26,41,56 * * * * /root/taoniu-go/cryptos binance futures tasks plans clea

* * * * * for i in $(seq 11) ; do /root/taoniu-go/cryptos binance margin cross tasks account flush ; sleep 5 ; done
```

# taoniu-config
淘牛配置
https://github.com/kuuy/taoniu-config

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
|现货短线交易        | binance spot tradings scalping                  |
|现货补仓交易        | binance spot tradings triggers                  |
|合约交易计划        | binance futures plans                              |
|合约短线交易        | binance futures tradings scalping                  |
|合约补仓交易        | binance futures tradings triggers                  |

# 免责声明
本项目仅为个人交易测试项目，风险意识完全靠自己把握，出现任何交易失误与本项目无关，请谨慎评估交易的合理性。

# 相关项目
[淘牛（react）](https://github.com/kuuy/taoniu-ts)  
[淘牛客户端（android）](https://github.com/kuuy/taoniu-android)  
[淘牛后台（react）](https://github.com/kuuy/taoniu-admin-ts)  
[淘牛后台（golang）](https://github.com/kuuy/taoniu-admin-go)  

![Anurag's GitHub stats](https://github-readme-stats.vercel.app/api?username=kuuy&show_icons=true&theme=radical)

![Codewars](https://github.r2v.ch/codewars?user=kuuy&top_languages=true&stroke=%23b362ff&theme=purple_dark)
