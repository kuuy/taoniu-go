package spot

import (
  "context"
  "fmt"
  "log"
  "math"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type PositionsHandler struct {
  Db                  *gorm.DB
  Rdb                 *redis.Client
  Ctx                 context.Context
  PositionsRepository *repositories.PositionsRepository
  SymbolsRepository   *repositories.SymbolsRepository
  TradingsRepository  *repositories.TradingsRepository
}

func NewPositionsCommand() *cli.Command {
  var h PositionsHandler
  return &cli.Command{
    Name:  "positions",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PositionsHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.TradingsRepository = &repositories.TradingsRepository{
        Db: h.Db,
      }
      h.TradingsRepository.ScalpingRepository = &tradingsRepositories.ScalpingRepository{
        Db: h.Db,
      }
      h.PositionsRepository = &repositories.PositionsRepository{
        Db: h.Db,
      }
      h.PositionsRepository.OrdersRepository = &repositories.OrdersRepository{
        Db:  h.Db,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "calc",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol can not be empty")
            return nil
          }
          margin, _ := strconv.ParseFloat(c.Args().Get(1), 16)
          leverage, _ := strconv.Atoi(c.Args().Get(2))
          entryPrice, _ := strconv.ParseFloat(c.Args().Get(3), 16)
          entryQuantity, _ := strconv.ParseFloat(c.Args().Get(4), 16)
          if err := h.Calc(symbol, margin, leverage, entryPrice, entryQuantity); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "apply",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if err := h.Apply(symbol); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if err := h.Flush(symbol); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "check",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Check(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *PositionsHandler) Calc(
  symbol string,
  margin float64,
  leverage int,
  entryPrice float64,
  entryQuantity float64,
) error {
  log.Println("binance spot positions calc...")

  maxCapital, _ := decimal.NewFromFloat(margin).Mul(decimal.NewFromInt32(int32(leverage))).Float64()
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  entity, err := h.SymbolsRepository.Get(symbol)
  if err != nil {
    return nil
  }

  var filters []string
  filters = strings.Split(entity.Filters["price"].(string), ",")
  tickSize, _ := strconv.ParseFloat(filters[2], 64)
  filters = strings.Split(entity.Filters["quote"].(string), ",")
  stepSize, _ := strconv.ParseFloat(filters[2], 64)

  entryQuantity, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryPrice)).Float64()
  log.Println("entry", entryPrice, strconv.FormatFloat(entryQuantity, 'f', -1, 64), entryAmount)

  var buyPrice float64
  var buyQuantity float64
  var buyAmount float64
  var sellPrice float64
  var takePrice float64

  if entryAmount < 5 {
    buyPrice = entryPrice
    buyQuantity = 5 / buyPrice
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
    entryQuantity = buyQuantity
    entryAmount = buyAmount
    sellPrice = h.PositionsRepository.SellPrice(entryPrice, entryAmount)
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    takePrice = h.PositionsRepository.TakePrice(entryPrice, tickSize)
  } else {
    takePrice = h.PositionsRepository.TakePrice(entryPrice, tickSize)
  }

  ipart, _ := math.Modf(maxCapital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  for {
    var err error
    capital, err := h.PositionsRepository.Capital(maxCapital, entryAmount, places)
    if err != nil {
      break
    }
    ratio := h.PositionsRepository.Ratio(capital, entryAmount)
    buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    if buyAmount < 5 {
      buyAmount = 5
    }
    buyQuantity = h.PositionsRepository.BuyQuantity(buyAmount, entryPrice, entryAmount)
    buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
    buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
    entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
    entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
    sellPrice = h.PositionsRepository.SellPrice(entryPrice, entryAmount)
    sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    log.Println("buy", buyPrice, strconv.FormatFloat(buyQuantity, 'f', -1, 64), buyAmount, sellPrice, entryPrice)
  }

  stopAmount, _ := decimal.NewFromFloat(entryAmount).Div(decimal.NewFromInt32(int32(leverage))).Mul(decimal.NewFromFloat(0.1)).Float64()

  var stopPrice float64
  stopPrice, _ = decimal.NewFromFloat(entryPrice).Sub(
    decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryQuantity)),
  ).Float64()
  stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()

  log.Println("takePrice", takePrice)
  log.Println("stopPrice", stopPrice)

  return nil
}

func (h *PositionsHandler) Apply(symbol string) error {
  var entities []*models.Scalping
  query := h.Db.Model(models.Scalping{}).Select([]string{"symbol"})
  if symbol != "" {
    query.Where("symbol", symbol)
  }
  query.Where("status", 1).Find(&entities)
  for _, entity := range entities {
    h.PositionsRepository.Apply(entity.Symbol)
    break
  }
  return nil
}

func (h *PositionsHandler) Flush(symbol string) error {
  var symbols []string
  if symbol == "" {
    symbols = h.TradingsRepository.Scan()
  } else {
    symbols = append(symbols, symbol)
  }
  for _, symbol := range symbols {
    mutex := common.NewMutex(
      h.Rdb,
      h.Ctx,
      fmt.Sprintf("locks:binance:spot:positions:flush:%s", symbol),
    )
    if !mutex.Lock(5 * time.Second) {
      return nil
    }
    defer mutex.Unlock()
    log.Println("symbol", symbol)
    if position, err := h.PositionsRepository.Get(symbol); err == nil {
      h.PositionsRepository.Flush(position)
    } else {
      h.PositionsRepository.Apply(symbol)
      log.Println("position not exists", symbol)
    }
    continue
  }
  return nil
}

func (h *PositionsHandler) Check() (err error) {
  log.Println("binance spot positions check...")
  symbols := h.TradingsRepository.Scan()
  for _, symbol := range symbols {
    position, _ := h.PositionsRepository.Get(symbol)
    entity, _ := h.SymbolsRepository.Get(symbol)
    value, _ := h.Rdb.HGet(h.Ctx, fmt.Sprintf("binance:spot:balance:%s", entity.BaseAsset), "free").Result()
    free, _ := strconv.ParseFloat(value, 64)
    if free < position.EntryQuantity {
      log.Println("balance free not enough", symbol, free, position.EntryQuantity)
    }
  }
  return
}
