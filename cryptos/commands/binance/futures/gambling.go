package futures

import (
  "fmt"
  "log"
  "strconv"

  "github.com/shopspring/decimal"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

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
          entryPrice, _ := strconv.ParseFloat(c.Args().Get(2), 64)
          entryQuantity, _ := strconv.ParseFloat(c.Args().Get(3), 64)
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

  entryPriceDec := decimal.NewFromFloat(entryPrice)
  entryQuantityDec := decimal.NewFromFloat(entryQuantity)
  entryAmountDec := entryPriceDec.Mul(entryQuantityDec)

  entity, err := h.SymbolsRepository.Get(symbol)
  if err != nil {
    return
  }

  tickSize, stepSize, notional, err := h.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    return
  }

  stepSizeDec := decimal.NewFromFloat(stepSize)
  notionalDec := decimal.NewFromFloat(notional)

  if stepSize > 0 {
    entryQuantityDec = entryAmountDec.Div(entryPriceDec).Div(stepSizeDec).Floor().Mul(stepSizeDec)
  } else {
    entryQuantityDec = entryAmountDec.Div(entryPriceDec)
  }
  entryQuantity, _ = entryQuantityDec.Float64()
  entryAmount, _ := entryAmountDec.Float64()

  log.Println(
    "entry",
    strconv.FormatFloat(entryPrice, 'f', -1, 64),
    strconv.FormatFloat(entryQuantity, 'f', -1, 64),
    entryAmount,
  )

  takePrice := h.GamblingRepository.TakePrice(entryPrice, side, tickSize)
  stopPrice := h.GamblingRepository.StopPrice(entryPrice, side, tickSize)
  takePriceDec := decimal.NewFromFloat(takePrice)

  planPrice := entryPrice
  planQuantityDec := entryQuantityDec
  planAmountDec := decimal.Zero
  planProfitDec := decimal.Zero
  lastProfitDec := decimal.Zero

  for {
    planPriceFloat, _ := decimal.NewFromFloat(planPrice).Float64()
    planQuantityFloat, _ := planQuantityDec.Float64()
    plans := h.GamblingRepository.Calc(planPriceFloat, planQuantityFloat, side, tickSize, stepSize)
    for _, plan := range plans {
      planTakeQuantityDec := decimal.NewFromFloat(plan.TakeQuantity)
      planTakePriceDec := decimal.NewFromFloat(plan.TakePrice)
      planTakeAmountDec := decimal.NewFromFloat(plan.TakeAmount)

      if planTakeQuantityDec.LessThan(stepSizeDec) {
        if side == 1 {
          lastProfitDec = takePriceDec.Sub(entryPriceDec).Mul(planQuantityDec)
        } else {
          lastProfitDec = entryPriceDec.Sub(takePriceDec).Mul(planQuantityDec)
        }
        break
      }
      if side == 1 && plan.TakePrice >= takePrice {
        lastProfitDec = takePriceDec.Sub(entryPriceDec).Mul(planQuantityDec)
        break
      }
      if side == 2 && plan.TakePrice <= takePrice {
        lastProfitDec = entryPriceDec.Sub(takePriceDec).Mul(planQuantityDec)
        break
      }

      var takeProfitDec decimal.Decimal
      if side == 1 {
        takeProfitDec = planTakePriceDec.Sub(entryPriceDec).Mul(planTakeQuantityDec)
      } else {
        takeProfitDec = entryPriceDec.Sub(planTakePriceDec).Mul(planTakeQuantityDec)
      }

      planPrice = plan.TakePrice
      planQuantityDec = planQuantityDec.Sub(planTakeQuantityDec)
      planAmountDec = planAmountDec.Add(planTakeAmountDec)
      planProfitDec = planProfitDec.Add(takeProfitDec)

      if planTakeAmountDec.LessThan(notionalDec) {
        return fmt.Errorf("plan amount less then %v", notional)
      }

      takeProfitFloat, _ := takeProfitDec.Float64()
      planAmountFloat, _ := planAmountDec.Float64()
      planProfitFloat, _ := planProfitDec.Float64()

      log.Println(
        "plan",
        strconv.FormatFloat(plan.TakePrice, 'f', -1, 64),
        strconv.FormatFloat(plan.TakeQuantity, 'f', -1, 64),
        takeProfitFloat,
        planAmountFloat,
        planProfitFloat,
      )
    }
    if len(plans) == 0 || lastProfitDec.GreaterThan(decimal.Zero) {
      break
    }
  }

  if planQuantityDec.GreaterThan(decimal.Zero) {
    var takeProfitDec decimal.Decimal
    if side == 1 {
      takeProfitDec = takePriceDec.Sub(entryPriceDec).Mul(planQuantityDec)
    } else {
      takeProfitDec = entryPriceDec.Sub(takePriceDec).Mul(planQuantityDec)
    }
    takeAmountDec := takePriceDec.Mul(planQuantityDec)
    planAmountDec = planAmountDec.Add(takeAmountDec)
    planProfitDec = planProfitDec.Add(takeProfitDec)

    if takeAmountDec.LessThan(notionalDec) {
      return fmt.Errorf("plan amount less then %v", notional)
    }

    planQuantityFloat, _ := planQuantityDec.Float64()
    takeProfitFloat, _ := takeProfitDec.Float64()
    planAmountFloat, _ := planAmountDec.Float64()
    planProfitFloat, _ := planProfitDec.Float64()

    log.Println(
      "plan",
      strconv.FormatFloat(takePrice, 'f', -1, 64),
      strconv.FormatFloat(planQuantityFloat, 'f', -1, 64),
      takeProfitFloat,
      planAmountFloat,
      planProfitFloat,
    )
  }

  planProfitFloat, _ := planProfitDec.Float64()

  log.Println("planProfit", planProfitFloat)
  log.Println("takePrice", strconv.FormatFloat(takePrice, 'f', -1, 64))
  log.Println("stopPrice", strconv.FormatFloat(stopPrice, 'f', -1, 64))

  return nil
}
