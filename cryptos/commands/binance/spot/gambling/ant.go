package gambling

import (
  "errors"
  "fmt"
  "log"
  "strconv"
  "strings"
  "time"

  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
  gamblingRepositories "taoniu.local/cryptos/repositories/binance/spot/gambling"
)

type AntHandler struct {
  Db                 *gorm.DB
  GamblingRepository *repositories.GamblingRepository
  AntRepository      *gamblingRepositories.AntRepository
  SymbolsRepository  *repositories.SymbolsRepository
}

func NewAntCommand() *cli.Command {
  var h AntHandler
  return &cli.Command{
    Name:  "ant",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AntHandler{
        Db: common.NewDB(1),
      }
      h.GamblingRepository = &repositories.GamblingRepository{
        Db: h.Db,
      }
      h.SymbolsRepository = &repositories.SymbolsRepository{
        Db: h.Db,
      }
      h.AntRepository = &gamblingRepositories.AntRepository{
        Db: h.Db,
      }
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
          entryPrice, _ := strconv.ParseFloat(c.Args().Get(1), 16)
          entryQuantity, _ := strconv.ParseFloat(c.Args().Get(2), 16)
          if err := h.Apply(symbol, entryPrice, entryQuantity); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
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

func (h *AntHandler) Apply(symbol string, entryPrice float64, entryQuantity float64) (err error) {
  log.Println("spot gambling ant apply...")

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

  antSide := 2

  takePrice := h.GamblingRepository.TakePrice(entryPrice, antSide, tickSize)

  planPrice := entryPrice
  planQuantity := entryQuantity
  lastPrice := 0.0
  lastProfit := 0.0

  var planPrices, planQuantities []float64
  var quantities []float64
  for {
    plans := h.GamblingRepository.Calc(planPrice, planQuantity, antSide, tickSize, stepSize)
    for _, plan := range plans {
      if plan.TakeQuantity < stepSize {
        if antSide == 1 {
          lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        } else {
          lastProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        }
        break
      }
      if antSide == 1 && plan.TakePrice > takePrice {
        lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        break
      }
      if antSide == 2 && plan.TakePrice < takePrice {
        lastProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
        break
      }
      planPrice = plan.TakePrice
      planQuantity, _ = decimal.NewFromFloat(planQuantity).Sub(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
      quantities = append(quantities, plan.TakeQuantity)
    }
    if len(plans) == 0 || lastProfit > 0 {
      break
    }
  }

  if planQuantity > 0 {
    quantities = append(quantities, planQuantity)
  }

  for _, lastQuantity := range quantities {
    planPrice = takePrice
    planQuantity = lastQuantity
    lastPrice = takePrice
    takePrice = h.GamblingRepository.TakePrice(lastPrice, antSide, tickSize)
    lastProfit = 0.0
    for {
      plans := h.GamblingRepository.Calc(planPrice, planQuantity, antSide, tickSize, stepSize)
      for _, plan := range plans {
        if plan.TakeQuantity < stepSize {
          if antSide == 1 {
            lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(lastPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          } else {
            lastProfit, _ = decimal.NewFromFloat(lastPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          }
          break
        }
        if antSide == 1 && plan.TakePrice > takePrice {
          lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(lastPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          break
        }
        if antSide == 2 && plan.TakePrice < takePrice {
          lastProfit, _ = decimal.NewFromFloat(lastPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          break
        }
        planPrice = plan.TakePrice
        planQuantity, _ = decimal.NewFromFloat(planQuantity).Sub(decimal.NewFromFloat(plan.TakeQuantity)).Float64()

        if plan.TakeAmount < notional {
          return errors.New(fmt.Sprintf("plan amount less then %v", notional))
        }

        planPrices = append(planPrices, plan.TakePrice)
        planQuantities = append(planQuantities, plan.TakeQuantity)
      }
      if len(plans) == 0 || lastProfit > 0 {
        break
      }
    }
    if planQuantity > 0 {
      takeAmount, _ := decimal.NewFromFloat(takePrice).Mul(decimal.NewFromFloat(planQuantity)).Float64()
      if takeAmount < notional {
        return errors.New(fmt.Sprintf("plan amount less then %v", notional))
      }

      planPrices = append(planPrices, takePrice)
      planQuantities = append(planQuantities, planQuantity)
    }
  }

  expiredAt := time.Now().Add(time.Hour * 24 * 14)
  err = h.AntRepository.Apply(symbol, entryPrice, entryQuantity, planPrices, planQuantities, expiredAt)
  if err != nil {
    return
  }

  return
}

func (h *AntHandler) Calc(
  symbol string,
  side int,
  entryPrice float64,
  entryQuantity float64,
) error {
  log.Println("binance spot positions calc...")

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

  takePrice := h.GamblingRepository.TakePrice(entryPrice, side, tickSize)

  planPrice := entryPrice
  planQuantity := entryQuantity
  planAmount := entryAmount
  planProfit := 0.0
  lastPrice := 0.0
  lastProfit := 0.0
  takeProfit := 0.0

  var quantities []float64
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
      planPrice = plan.TakePrice
      planQuantity, _ = decimal.NewFromFloat(planQuantity).Sub(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
      quantities = append(quantities, plan.TakeQuantity)
    }
    if len(plans) == 0 || lastProfit > 0 {
      break
    }
  }

  if planQuantity > 0 {
    planAmount = 0.0
    quantities = append(quantities, planQuantity)
  }

  for _, lastQuantity := range quantities {
    planPrice = takePrice
    planQuantity = lastQuantity
    lastPrice = takePrice
    takePrice = h.GamblingRepository.TakePrice(lastPrice, side, tickSize)
    lastProfit = 0.0
    for {
      plans := h.GamblingRepository.Calc(planPrice, planQuantity, side, tickSize, stepSize)
      for _, plan := range plans {
        if plan.TakeQuantity < stepSize {
          if side == 1 {
            lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(lastPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          } else {
            lastProfit, _ = decimal.NewFromFloat(lastPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          }
          break
        }
        if side == 1 && plan.TakePrice > takePrice {
          lastProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(lastPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          break
        }
        if side == 2 && plan.TakePrice < takePrice {
          lastProfit, _ = decimal.NewFromFloat(lastPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
          break
        }
        if side == 1 {
          takeProfit, _ = decimal.NewFromFloat(plan.TakePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
        } else {
          takeProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(plan.TakePrice)).Mul(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
        }
        planPrice = plan.TakePrice
        planQuantity, _ = decimal.NewFromFloat(planQuantity).Sub(decimal.NewFromFloat(plan.TakeQuantity)).Float64()
        planAmount, _ = decimal.NewFromFloat(planAmount).Add(decimal.NewFromFloat(plan.TakeAmount)).Float64()
        planProfit, _ = decimal.NewFromFloat(planProfit).Add(decimal.NewFromFloat(takeProfit)).Float64()
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
      planAmount, _ = decimal.NewFromFloat(planAmount).Add(decimal.NewFromFloat(takePrice).Mul(decimal.NewFromFloat(planQuantity))).Float64()
      planProfit, _ = decimal.NewFromFloat(planProfit).Add(decimal.NewFromFloat(takeProfit)).Float64()
      log.Println("plan", takePrice, planQuantity, takeProfit, planAmount, planProfit)
    }
  }

  log.Println("planProfit", planProfit)
  log.Println("takePrice", takePrice)

  return nil
}
