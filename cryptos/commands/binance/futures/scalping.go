package futures

import (
  "context"
  "fmt"
  "log"
  "strconv"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  config "taoniu.local/cryptos/config/binance/futures"
  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
  indicators "taoniu.local/cryptos/repositories/binance/futures/indicators"
)

type ScalpingHandler struct {
  Db                 *gorm.DB
  Rdb                *redis.Client
  Ctx                context.Context
  ScalpingRepository *repositories.ScalpingRepository
  SymbolsRepository  *repositories.SymbolsRepository
  StopLossRepository *repositories.StopLossRepository
}

func NewScalpingCommand() *cli.Command {
  var h ScalpingHandler
  return &cli.Command{
    Name:  "scalping",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ScalpingHandler{
        Db:  common.NewDB(2),
        Rdb: common.NewRedis(2),
        Ctx: context.Background(),
      }
      h.ScalpingRepository = &repositories.ScalpingRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db:  h.Db,
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.StopLossRepository = &repositories.StopLossRepository{
        Db:                h.Db,
        Rdb:               h.Rdb,
        Ctx:               h.Ctx,
        SymbolsRepository: h.SymbolsRepository,
      }
      h.StopLossRepository.TickersRepository = &repositories.TickersRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      h.StopLossRepository.AtrRepository = &indicators.AtrRepository{}
      h.StopLossRepository.AtrRepository.Db = h.Db
      h.StopLossRepository.AtrRepository.Rdb = h.Rdb
      h.StopLossRepository.AtrRepository.Ctx = h.Ctx
      h.StopLossRepository.VolumeProfileRepository = &indicators.VolumeProfileRepository{}
      h.StopLossRepository.VolumeProfileRepository.Db = h.Db
      h.StopLossRepository.VolumeProfileRepository.Rdb = h.Rdb
      h.StopLossRepository.VolumeProfileRepository.Ctx = h.Ctx
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "apply",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol can not be empty")
            return nil
          }
          side, _ := strconv.Atoi(c.Args().Get(1))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Apply(symbol, side); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "plans",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Plans(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          side, _ := strconv.Atoi(c.Args().Get(0))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Flush(side); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "check",
        Usage: "",
        Action: func(c *cli.Context) error {
          side, _ := strconv.Atoi(c.Args().Get(0))
          if side != 1 && side != 2 {
            log.Fatal("side invalid")
            return nil
          }
          if err := h.Check(side); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "copy",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Copy(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "stoploss-calc",
        Usage: "Calculate stop loss price",
        Flags: []cli.Flag{
          &cli.StringFlag{
            Name:     "symbol",
            Aliases:  []string{"s"},
            Usage:    "trading symbol (e.g. BTCUSDT)",
            Required: true,
          },
          &cli.IntFlag{
            Name:     "side",
            Usage:    "trading side (1 for Long, 2 for Short)",
            Required: true,
          },
          &cli.Float64Flag{
            Name:     "entry",
            Aliases:  []string{"e"},
            Usage:    "entry price",
            Required: true,
          },
          &cli.IntFlag{
            Name:    "leverage",
            Aliases: []string{"l"},
            Usage:   "leverage",
            Value:   10,
          },
          &cli.Float64Flag{
            Name:    "risk",
            Aliases: []string{"r"},
            Usage:   "risk ratio (e.g. 0.02 for 2%)",
            Value:   0.02,
          },
        },
        Action: func(c *cli.Context) error {
          symbol := c.String("symbol")
          side := c.Int("side")
          entry := c.Float64("entry")
          leverage := c.Int("leverage")
          risk := c.Float64("risk")
          if err := h.StopLossCalc(symbol, side, entry, leverage, risk); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ScalpingHandler) Apply(symbol string, side int) error {
  log.Println("futures scalping apply...")

  capital := 5000.0

  data, _ := h.Rdb.HMGet(
    h.Ctx,
    fmt.Sprintf(
      "binance:futures:indicators:1d:%s:%s",
      symbol,
      time.Now().Format("0102"),
    ),
    "take_profit_price",
    "stop_loss_point",
  ).Result()
  if data[0] == nil || data[1] == nil {
    data, _ = h.Rdb.HMGet(
      h.Ctx,
      fmt.Sprintf(
        "binance:futures:indicators:1m:%s:%s",
        symbol,
        time.Now().Format("0102"),
      ),
      "take_profit_price",
      "stop_loss_point",
    ).Result()
    if data[0] == nil || data[1] == nil {
      return fmt.Errorf("[%s] indicators empty", symbol)
    }
  }
  takePrice, _ := strconv.ParseFloat(data[0].(string), 64)
  stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

  var price float64
  if side == 1 {
    price = takePrice
  } else {
    price = stopPrice
  }

  expiredAt := time.Now().Add(time.Hour * 24 * 14)
  err := h.ScalpingRepository.Apply(symbol, side, capital, price, expiredAt)
  if err != nil {
    return err
  }

  return nil
}

func (h *ScalpingHandler) Plans() error {
  log.Println("futures scalping plans...")
  var plans []*models.Plan
  h.Db.Model(&plans).Select([]string{"symbol", "count(1) as times"}).Group("symbol").Having("count(1) > 10").Order("times desc").Find(&plans)
  for _, plan := range plans {
    h.Apply(plan.Symbol, plan.Side)
  }
  return nil
}

func (h *ScalpingHandler) Flush(side int) error {
  var scalping []*models.Scalping
  h.Db.Model(&models.Scalping{}).Where("side=? AND status in (1,2)", side).Find(&scalping)
  for _, entity := range scalping {
    data, _ := h.Rdb.HMGet(
      h.Ctx,
      fmt.Sprintf(
        "binance:futures:indicators:1d:%s:%s",
        entity.Symbol,
        time.Now().Format("0102"),
      ),
      "vah",
      "val",
    ).Result()
    if len(data) == 0 || data[0] == nil || data[1] == nil {
      log.Println("indicators empty", entity.Symbol)
      continue
    }

    takePrice, _ := strconv.ParseFloat(data[0].(string), 64)
    stopPrice, _ := strconv.ParseFloat(data[1].(string), 64)

    var price float64
    if side == 1 {
      price = takePrice
    } else {
      price = stopPrice
    }

    log.Println("scalping update", entity.Symbol, entity.Side, price)

    h.Db.Model(&entity).Update("price", price)
  }
  return nil
}

func (h *ScalpingHandler) Check(side int) error {
  var scalping []*models.Scalping
  h.Db.Model(&models.Scalping{}).Where("side=? AND status in (1,2)", side).Find(&scalping)

  now := time.Now().UTC()
  duration := -time.Second * time.Duration(now.Second())
  duration = duration - time.Hour*time.Duration(now.Hour()) - time.Minute*time.Duration(now.Minute())
  timestamp := now.Add(duration).Unix() * 1000

  for _, entity := range scalping {
    redisKey := fmt.Sprintf(config.REDIS_KEY_KLINES, "1d", entity.Symbol, timestamp)
    exists, _ := h.Rdb.Exists(h.Ctx, redisKey).Result()
    if exists != 1 {
      log.Println("scalping klines not exists", entity.Symbol)
    }
  }

  for _, entity := range scalping {
    item, err := h.SymbolsRepository.Get(entity.Symbol)
    if err != nil {
      return err
    }
    _, stepSize, notional, err := h.SymbolsRepository.Filters(item.Filters)
    if err != nil {
      return nil
    }
    price, err := h.SymbolsRepository.Price(item.Symbol)
    if err != nil {
      return err
    }
    buyQuantity, _ := decimal.NewFromFloat(notional).Div(decimal.NewFromFloat(price)).Float64()
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    if buyQuantity*price >= 10.0 {
      println("scalping", entity.Symbol, fmt.Sprintf("%v %v", notional, buyQuantity*price))
    }
  }

  return nil
}

func (h *ScalpingHandler) Copy() error {
  var scalping []*models.Scalping
  h.Db.Model(&models.Scalping{}).Where("side=? AND status=1", 2).Find(&scalping)
  for _, entity := range scalping {
    h.Apply(entity.Symbol, 1)
  }
  return nil
}

func (h *ScalpingHandler) Init() error {
  var symbols []string
  h.Db.Model(&models.Kline{}).Select("DISTINCT symbol").Where("interval=?", "1d").Find(&symbols)
  for _, symbol := range symbols {
    h.Apply(symbol, 2)
  }
  return nil
}

func (h *ScalpingHandler) StopLossCalc(symbol string, side int, entry float64, leverage int, risk float64) error {
  result, err := h.StopLossRepository.Calc(symbol, side, entry, 0, leverage, risk)
  if err != nil {
    return err
  }

  var sideStr string
  if result.Side == 1 {
    sideStr = "LONG"
  } else {
    sideStr = "SHORT"
  }

  fmt.Printf("Symbol:         %s\n", result.Symbol)
  fmt.Printf("Side:           %s\n", sideStr)
  fmt.Printf("Entry:          %v\n", result.EntryPrice)
  fmt.Printf("Leverage:       %dx\n", result.Leverage)
  fmt.Printf("Risk:           %.2f%%\n", result.Risk*100)
  fmt.Printf("ATR:            %v (Multiplier: %v)\n", result.ATR, result.ATRMultiplier)
  fmt.Printf("Initial Stop:   %v\n", result.InitialStop)
  fmt.Printf("Take Profit 1:  %v\n", result.TakeProfit1)
  fmt.Printf("Take Profit 2:  %v\n", result.TakeProfit2)
  fmt.Printf("Risk/Reward:    %v\n", result.RiskReward)
  fmt.Printf("Should Trade:   %v\n", result.ShouldTrade)
  if result.RejectReason != "" {
    fmt.Printf("Reject Reason:  %s\n", result.RejectReason)
  }

  return nil
}
