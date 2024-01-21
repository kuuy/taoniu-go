package spot

import (
  "context"
  "errors"
  "fmt"
  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"
  "os"
  "slices"
  "strconv"
)

type AccountRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *AccountRepository) Flush() error {
  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"),
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  account, err := client.NewGetAccountService().Do(r.Ctx)
  if err != nil {
    return err
  }
  oldCurrencies, _ := r.Rdb.SMembers(r.Ctx, "binance:spot:currencies").Result()
  var currencies []string
  for _, coin := range account.Balances {
    free, _ := strconv.ParseFloat(coin.Free, 64)
    if free <= 0.0 {
      r.Rdb.SRem(r.Ctx, "binance:spot:currencies", coin.Asset)
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:balance:%s", coin.Asset))
      continue
    }
    r.Rdb.SAdd(r.Ctx, "binance:spot:currencies", coin.Asset)
    r.Rdb.HMSet(
      r.Ctx,
      fmt.Sprintf("binance:spot:balance:%s", coin.Asset),
      map[string]interface{}{
        "free":   coin.Free,
        "locked": coin.Locked,
      },
    )
    currencies = append(currencies, coin.Asset)
  }
  for _, currency := range oldCurrencies {
    if !slices.Contains(currencies, currency) {
      r.Rdb.SRem(r.Ctx, "binance:spot:currencies", currency)
      r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:balance:%s", currency))
    }
  }

  return nil
}

func (r *AccountRepository) Balance(asset string) (map[string]float64, error) {
  fields := []string{
    "free",
    "locked",
  }
  data, _ := r.Rdb.HMGet(
    r.Ctx,
    fmt.Sprintf(
      "binance:spot:balance:%s",
      asset,
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
