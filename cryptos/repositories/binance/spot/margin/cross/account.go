package cross

import (
  "context"
  "errors"
  "fmt"
  "strconv"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  config "taoniu.local/cryptos/config/binance/spot"
)

type AccountRepository struct {
  Rdb *redis.Client
  Ctx context.Context
}

func (r *AccountRepository) Flush() error {
  client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
  account, err := client.NewGetMarginAccountService().Do(r.Ctx)
  if err != nil {
    return err
  }
  for _, coin := range account.UserAssets {
    netAsset, _ := strconv.ParseFloat(coin.NetAsset, 64)
    if netAsset == 0.0 {
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:margin:cross:balance:%s", coin.Asset))
      continue
    }
    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf("binance:spot:margin:cross:balance:%s", coin.Asset),
      map[string]interface{}{
        "free":      coin.Free,
        "locked":    coin.Locked,
        "borrowed":  coin.Borrowed,
        "interest":  coin.Interest,
        "net_asset": coin.NetAsset,
      },
    )
  }

  return nil
}

func (r *AccountRepository) Balance(symbol string) (map[string]float64, error) {
  fields := []string{
    "free",
    "locked",
    "borrowed",
    "interest",
    "net_asset",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:spot:margin:cross:balance:%s",
      symbol,
    ),
    fields...,
  ).Result()
  balance := map[string]float64{}
  for i, field := range fields {
    if data[i] == nil {
      return nil, errors.New("balance not exists")
    }
    balance[field], _ = strconv.ParseFloat(data[i].(string), 64)
  }
  return balance, nil
}
