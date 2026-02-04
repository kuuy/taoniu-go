package futures

import (
  "fmt"
  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "strconv"

  "taoniu.local/cryptos/commands/binance/futures/gambling"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type GamblingHandler struct {
  Db                 *gorm.DB
  GamblingRepository *repositories.GamblingRepository
  SymbolsRepository  *repositories.SymbolsRepository
}

func NewGamblingCommand() *cli.Command {
  var h GamblingHandler
  return &cli.Command{
    Name:  "gambling",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = GamblingHandler{
        Db: common.NewDB(2),
      }
      h.GamblingRepository = &repositories.GamblingRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      gambling.NewAntCommand(),
      {
        Name:  "calc",
        Usage: "",
        Action: func(c *cli.Context) error {
          symbol := c.Args().Get(0)
          if symbol == "" {
            log.Fatal("symbol can not be empty")
            return nil
          }
          side, _ := strconv.Atoi(c.Args().Get(1))
          entryPrice, _ := strconv.ParseFloat(c.Args().Get(2), 16)
          entryQuantity, _ := strconv.ParseFloat(c.Args().Get(3), 16)
          if err := h.Calc(symbol, side, entryPrice, entryQuantity); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *GamblingHandler) Calc(
  symbol string,
  side int,
  entryPrice float64,
  entryQuantity float64,
) (err error) {
  log.Println("binance futures positions calc...")

  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  entity, err := h.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, notional, err := h.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  entryQuantity, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryPrice)).Float64()
  log.Println("entry", entryPrice, strconv.FormatFloat(entryQuantity, 'f', -1, 64), entryAmount)

  takePrice := h.GamblingRepository.TakePrice(entryPrice, side, tickSize)
  stopPrice := h.GamblingRepository.StopPrice(entryPrice, side, tickSize)

  planPrice := entryPrice
  planQuantity := entryQuantity
  planAmount := entryAmount
  planProfit := 0.0
  lastProfit := 0.0
  takeProfit := 0.0

  for {
    plans := h.GamblingRepository.Calc(planPrice, planQuantity, side, tickSize, stepSize)
    for _, plan := range plans {
      if plan.TakeQuantity < stepSize {
        if side == 1 {
          lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        } else {
          lastProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        }
        break
      }
      if side == 1 && plan.TakePrice > takePrice {
        lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        break
      }
      if side == 2 && plan.TakePrice < takePrice {
        lastProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        break
      }
      if side == 1 {
        takeProfit, _ = decimal.NewFromFloat(plan.TakePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
      } else {
        takeProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(plan.TakePrice)).Mul(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
      }
      planPrice = plan.TakePrice
      planQuantity, _ = decimal.NewFromFloat(planQuantity).Sub(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
      planAmount, _ = decimal.NewFromFloat(planAmount).Sub(decimal.NewFromFloat(plan.TakeAmount)).Float64()
      planProfit, _ = decimal.NewFromFloat(planProfit).Add(decimal.NewFromFloat(takeProfit)).Float64()

      if plan.TakeAmount < notional {
        return fmt.Errorf("plan amount less then %v", notional)
      }

      log.Println("plan", plan.TakePrice, strconv.FormatFloat(plan.TakeQuantity, 'f', -1, 64), takeProfit, planAmount, planProfit)
    }
    if len(plans) == 0 || lastProfit > 0 {
      break
    }
  }

  if planQuantity > 0 {
    if side == 1 {
      takeProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
    } else {
      takeProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
    }
    takeAmount, _ := decimal.NewFromFloat(takePrice).Mul(decimal.NewFromFloat(planQuantity)).Float64()
    planAmount, _ = decimal.NewFromFloat(planAmount).Add(decimal.NewFromFloat(takeAmount)).Float64()
    planProfit, _ = decimal.NewFromFloat(planProfit).Add(decimal.NewFromFloat(takeProfit)).Float64()

    if takeAmount < notional {
      return fmt.Errorf("plan amount less then %v", notional)
    }

    log.Println("plan", takePrice, planQuantity, takeProfit, planAmount, planProfit)
  }

  log.Println("planProfit", planProfit)
  log.Println("takePrice", takePrice)
  log.Println("stopPrice", stopPrice)

  return nil
}
