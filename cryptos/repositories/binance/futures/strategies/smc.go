package strategies

import (
  "context"
  "fmt"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  models "taoniu.local/cryptos/models/binance/futures"
)

type SMCRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

type MarketStructure struct {
  Trend    int // 1: Bullish, 2: Bearish
  LastHigh float64
  LastLow  float64
  BOS      bool
  CHoCH    bool
}

type OrderBlock struct {
  Price       float64
  High        float64
  Low         float64
  Volume      float64
  Side        int // 1: Bullish OB, 2: Bearish OB
  IsMitigated bool
}

type FVG struct {
  Top    float64
  Bottom float64
  Side   int // 1: Bullish, 2: Bearish
}

// IdentifyMarketStructure 识别市场结构 (BOS/CHoCH)
func (r *SMCRepository) IdentifyMarketStructure(klines []*models.Kline) *MarketStructure {
  if len(klines) < 20 {
    return nil
  }

  ms := &MarketStructure{}
  // 这里简化实现：使用最后几个波峰波谷
  // 实际生产中建议使用 ZigZag 算法先提取波点

  // 简单的趋势判定：收盘价在 20 均线上方为多
  var sum float64
  for i := 0; i < 20; i++ {
    sum += klines[i].Close
  }
  ma := sum / 20
  if klines[0].Close > ma {
    ms.Trend = 1
  } else {
    ms.Trend = 2
  }

  // 识别 CHoCH (趋势改变)
  // 如果是多头趋势，但价格跌破了前一个显著低点 -> CHoCH
  if ms.Trend == 1 && klines[0].Close < klines[5].Low {
    ms.CHoCH = true
  }
  // 识别 BOS (趋势延续)
  // 如果是多头趋势，且价格突破了前一个显著高点 -> BOS
  if ms.Trend == 1 && klines[0].High > klines[5].High {
    ms.BOS = true
  }

  return ms
}

// FindOrderBlocks 寻找订单块 (Order Blocks)
// 逻辑：在一波强力上涨/下跌之前的最后一根反向阴线/阳线
func (r *SMCRepository) FindOrderBlocks(klines []*models.Kline) []*OrderBlock {
  var obs []*OrderBlock
  for i := 1; i < len(klines)-5; i++ {
    // 寻找 Bullish OB: 强力上涨前的最后一根阴线
    if klines[i].Close < klines[i].Open { // 阴线
      // 检查后面是否有强力上涨 (连续 3 根阳线且放量)
      if klines[i-1].Close > klines[i-1].Open && klines[i-2].Close > klines[i-2].Open {
        if (klines[i-2].Close-klines[i].Close)/klines[i].Close > 0.01 { // 1% 以上涨幅
          obs = append(obs, &OrderBlock{
            Price:  klines[i].Close,
            High:   klines[i].High,
            Low:    klines[i].Low,
            Volume: klines[i].Volume,
            Side:   1,
          })
        }
      }
    }
    // 寻找 Bearish OB: 强力下跌前的最后一根阳线
    if klines[i].Close > klines[i].Open { // 阳线
      if klines[i-1].Close < klines[i-1].Open && klines[i-2].Close < klines[i-2].Open {
        if (klines[i].Close-klines[i-2].Close)/klines[i].Close > 0.01 {
          obs = append(obs, &OrderBlock{
            Price:  klines[i].Close,
            High:   klines[i].High,
            Low:    klines[i].Low,
            Volume: klines[i].Volume,
            Side:   2,
          })
        }
      }
    }
  }
  return obs
}

// FindFVG 寻找公允价值缺口 (Fair Value Gap)
// 逻辑：3 根 K 线模型，第 1 根的高点与第 3 根的低点之间存在未重合区域
func (r *SMCRepository) FindFVG(klines []*models.Kline) []*FVG {
  var fvgs []*FVG
  for i := 0; i < len(klines)-3; i++ {
    // Bullish FVG
    if klines[i+2].High < klines[i].Low {
      fvgs = append(fvgs, &FVG{
        Top:    klines[i].Low,
        Bottom: klines[i+2].High,
        Side:   1,
      })
    }
    // Bearish FVG
    if klines[i+2].Low > klines[i].High {
      fvgs = append(fvgs, &FVG{
        Top:    klines[i+2].Low,
        Bottom: klines[i].High,
        Side:   2,
      })
    }
  }
  return fvgs
}

// Scan 扫描 SMC 机会
func (r *SMCRepository) Scan(symbol string, interval string) (signal int, price float64, err error) {
  var klines []*models.Kline
  r.Db.Where("symbol=? AND interval=?", symbol, interval).Order("timestamp desc").Limit(100).Find(&klines)

  if len(klines) < 50 {
    return 0, 0, fmt.Errorf("klines not enough")
  }

  ms := r.IdentifyMarketStructure(klines)
  obs := r.FindOrderBlocks(klines)
  fvgs := r.FindFVG(klines)

  currentPrice := klines[0].Close

  // SMC 交易核心逻辑示例：
  // 1. 结构看多 (BOS 或 多头趋势)
  // 2. 价格回调至下方的 Bullish Order Block 或 填补 Bullish FVG
  // 3. 产生做多信号

  if ms.Trend == 1 {
    for _, ob := range obs {
      if ob.Side == 1 && currentPrice > ob.Low && currentPrice < ob.High {
        // 价格触及看多订单块
        return 1, currentPrice, nil
      }
    }
    for _, f := range fvgs {
      if f.Side == 1 && currentPrice < f.Top && currentPrice > f.Bottom {
        // 价格进入看多缺口
        return 1, currentPrice, nil
      }
    }
  }

  // 做空逻辑同理
  if ms.Trend == 2 {
    for _, ob := range obs {
      if ob.Side == 2 && currentPrice < ob.High && currentPrice > ob.Low {
        return 2, currentPrice, nil
      }
    }
  }

  return 0, 0, nil
}
