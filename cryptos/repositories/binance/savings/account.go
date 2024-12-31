package savings

import (
  "context"
  "log"
  "os"

  "github.com/adshao/go-binance/v2"
  "github.com/go-redis/redis/v8"
)

type AccountRepository struct {
  Rdb *redis.Client
  Ctx context.Context
}

func (r *AccountRepository) Flush() error {
  client := binance.NewClient(
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_KEY"),
    os.Getenv("BINANCE_SPOT_ACCOUNT_API_SECRET"),
  )
  client.BaseURL = os.Getenv("BINANCE_SPOT_API_ENDPOINT")

  products, err := client.NewSavingFixedProjectPositionsService().Status("HOLDING").Do(r.Ctx)
  if err != nil {
    log.Println("error:", err)
    return err
  }
  log.Println("products:", products)
  for _, product := range products {
    log.Println("product:", product)
    //	free, _ := strconv.ParseFloat(coin.Free, 64)
    //	if free <= 0.0 {
    //		r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:balance:%s", coin.Asset))
    //		continue
    //	}
    //	r.Rdb.HMSet(
    //		r.Ctx,
    //		fmt.Sprintf("binance:spot:balance:%s", coin.Asset),
    //		map[string]interface{}{
    //			"free":   coin.Free,
    //			"locked": coin.Locked,
    //		},
    //	)
  }

  return nil
}
