package futures

import (
  "log"
  "math"
  "strconv"
  "strings"

  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type PositionsHandler struct {
  Db                *gorm.DB
  Repository        *repositories.PositionsRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewPositionsCommand() *cli.Command {
  var h PositionsHandler
  return &cli.Command{
    Name:  "positions",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = PositionsHandler{
        Db: common.NewDB(2),
      }
      h.Repository = &repositories.PositionsRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
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
          side, _ := strconv.Atoi(c.Args().Get(3))
          entryPrice, _ := strconv.ParseFloat(c.Args().Get(4), 16)
          entryQuantity, _ := strconv.ParseFloat(c.Args().Get(5), 16)
          if err := h.Calc(symbol, margin, leverage, side, entryPrice, entryQuantity); err != nil {
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
          side, _ := strconv.Atoi(c.Args().Get(1))
          if err := h.Flush(symbol, side); err != nil {
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
  side int,
  entryPrice float64,
  entryQuantity float64,
) error {
  log.Println("binance futures positions calc...")

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
  log.Println("entry", entryPrice, entryQuantity, entryAmount)

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
    sellPrice = h.Repository.SellPrice(side, entryPrice, entryAmount)
    if side == 1 {
      sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    } else {
      sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    }
    takePrice = h.Repository.TakePrice(entryPrice, side, tickSize)
  } else {
    takePrice = h.Repository.TakePrice(entryPrice, side, tickSize)
  }

  ipart, _ := math.Modf(maxCapital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  for {
    var err error
    capital, err := h.Repository.Capital(maxCapital, entryAmount, places)
    if err != nil {
      break
    }
    ratio := h.Repository.Ratio(capital, entryAmount)
    buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    if buyAmount < 5 {
      buyAmount = 5
    }
    buyQuantity = h.Repository.BuyQuantity(side, buyAmount, entryPrice, entryAmount)
    buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
    if side == 1 {
      buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    } else {
      buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    }
    buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
    buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
    entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
    entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
    sellPrice = h.Repository.SellPrice(side, entryPrice, entryAmount)
    if side == 1 {
      sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    } else {
      sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    }
    log.Println("buy", buyPrice, buyQuantity, buyAmount, sellPrice, entryPrice)
  }

  stopAmount, _ := decimal.NewFromFloat(entryAmount).Div(decimal.NewFromInt32(int32(leverage))).Mul(decimal.NewFromFloat(0.1)).Float64()

  var stopPrice float64
  if side == 1 {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Sub(
      decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryQuantity)),
    ).Float64()
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Add(
      decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryQuantity)),
    ).Float64()
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  log.Println("takePrice", takePrice)
  log.Println("stopPrice", stopPrice)

  return nil
}

func (h *PositionsHandler) Flush(symbol string, side int) error {
  var triggers []*models.Trigger
  query := h.Db.Model(models.Trigger{}).Select([]string{"symbol", "side"})
  if symbol != "" {
    query.Where("symbol", symbol)
  }
  if side == 1 || side == 2 {
    query.Where("side", side)
  }
  query.Where("status", 1).Find(&triggers)
  for _, trigger := range triggers {
    h.Repository.Flush(trigger.Symbol, trigger.Side)
    break
  }
  return nil
}
