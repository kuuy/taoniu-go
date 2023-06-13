package futures

import (
  "log"
  "strconv"
  "strings"

  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
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
        Db: common.NewDB(),
      }
      h.Repository = &repositories.PositionsRepository{}
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
          multiple, _ := strconv.Atoi(c.Args().Get(2))
          side, _ := strconv.Atoi(c.Args().Get(3))
          entryPrice, _ := strconv.ParseFloat(c.Args().Get(4), 16)
          entryVolume, _ := strconv.ParseFloat(c.Args().Get(5), 16)
          if err := h.calc(symbol, margin, multiple, side, entryPrice, entryVolume); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *PositionsHandler) calc(
  symbol string,
  margin float64,
  multiple int,
  side int,
  entryPrice float64,
  entryVolume float64,
) error {
  log.Println("binance futures positions calc...")

  capital, _ := decimal.NewFromFloat(margin).Mul(decimal.NewFromInt32(int32(multiple))).Float64()
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryVolume)).Float64()

  entity, err := h.SymbolsRepository.Get(symbol)
  if err != nil {
    return nil
  }

  var filters []string
  filters = strings.Split(entity.Filters["price"].(string), ",")
  tickSize, _ := strconv.ParseFloat(filters[2], 64)
  filters = strings.Split(entity.Filters["quote"].(string), ",")
  stepSize, _ := strconv.ParseFloat(filters[2], 64)

  ratio := h.Repository.Ratio(capital, entryAmount)
  if entryAmount == 0.0 {
    entryAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
    ratio = h.Repository.Ratio(capital, entryAmount)
  }
  entryVolume, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryPrice)).Float64()
  log.Println("entry", entryPrice, entryVolume, entryAmount)

  var takePrice float64
  if side == 1 {
    takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.02)).Float64()
    takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    takePrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.98)).Float64()
    takePrice, _ = decimal.NewFromFloat(takePrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }
  log.Println("takePrice", takePrice)

  for ratio > 0.0 {
    if side == 1 {
      entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(0.995)).Float64()
    } else {
      entryPrice, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(1.005)).Float64()
    }
    price, volume, amount := h.Repository.Calc(capital, side, entryPrice, entryAmount, ratio)
    if side == 1 {
      price, _ = decimal.NewFromFloat(price).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
    } else {
      price, _ = decimal.NewFromFloat(price).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
    }
    volume, _ = decimal.NewFromFloat(volume).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()

    entryVolume, _ = decimal.NewFromFloat(entryVolume).Add(decimal.NewFromFloat(volume)).Float64()
    entryAmount, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryVolume)).Float64()

    log.Println("price", price, volume, amount, entryPrice)
    ratio = h.Repository.Ratio(capital, entryAmount)
  }

  stopAmount, _ := decimal.NewFromFloat(margin).Mul(decimal.NewFromFloat(0.1)).Float64()

  var stopPrice float64
  if side == 1 {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Sub(
      decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryVolume)),
    ).Float64()
  } else {
    stopPrice, _ = decimal.NewFromFloat(entryPrice).Add(
      decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryVolume)),
    ).Float64()
  }

  if side == 1 {
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
  } else {
    stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
  }

  log.Println("stopPrice", stopPrice)

  return nil
}
