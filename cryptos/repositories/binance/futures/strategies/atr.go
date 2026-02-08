package strategies

import (
  "fmt"
  "math"
  "time"

  config "taoniu.local/cryptos/config/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type AtrRepository struct {
  BaseRepository
  Repository *repositories.AtrRepository
}

func (r *AtrRepository) Flush(symbol string, interval string) (err error) {
  tickSize, _, err := r.Filters(symbol)
  if err != nil {
    return err
  }

  day := time.Now().Format("0102")
  atr, err := r.Repository.Get(symbol, interval)
  if err != nil {
    return
  }
  price, err := r.Price(symbol)
  if err != nil {
    return
  }
  profitTarget := math.Ceil((price*2-atr*1.5)/tickSize) * tickSize
  stopLossPoint := math.Floor((price-atr)/tickSize) * tickSize
  takeProfitPrice := math.Ceil((stopLossPoint+(profitTarget-stopLossPoint)/2)/tickSize) * tickSize
  riskRewardRatio := math.Round((price-stopLossPoint)/(profitTarget-price)*100) / 100
  takeProfitRatio := math.Round(price/takeProfitPrice*100) / 100

  r.Rdb.HMSet(
    r.Ctx,
    fmt.Sprintf(
      config.REDIS_KEY_INDICATORS,
      interval,
      symbol,
      day,
    ),
    map[string]interface{}{
      "profit_target":     profitTarget,
      "stop_loss_point":   stopLossPoint,
      "take_profit_price": takeProfitPrice,
      "risk_reward_ratio": riskRewardRatio,
      "take_profit_ratio": takeProfitRatio,
    },
  )

  return
}
