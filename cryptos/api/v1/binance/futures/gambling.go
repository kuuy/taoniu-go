package futures

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  "github.com/shopspring/decimal"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type GamblingHandler struct {
  ApiContext          *common.ApiContext
  Response            *api.ResponseHandler
  GamblingRepository  *repositories.GamblingRepository
  PositionsRepository *repositories.PositionsRepository
  SymbolsRepository   *repositories.SymbolsRepository
}

func NewGamblingRouter(apiContext *common.ApiContext) http.Handler {
  h := GamblingHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.Jwe = &common.Jwe{}
  h.GamblingRepository = &repositories.GamblingRepository{
    Db: h.ApiContext.Db,
  }
  h.PositionsRepository = &repositories.PositionsRepository{
    Db: h.ApiContext.Db,
  }
  h.SymbolsRepository = &repositories.SymbolsRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/calc", h.Calc)

  return r
}

func (h *GamblingHandler) Calc(
  w http.ResponseWriter,
  r *http.Request,
) {
  q := r.URL.Query()

  symbol := q.Get("symbol")
  if symbol == "" {
    h.Response.Error(w, http.StatusForbidden, 1004, "symbol is empty")
    return
  }

  side := 1
  if q.Get("side") != "" {
    side, _ = strconv.Atoi(q.Get("side"))
  }

  var entryPrice float64
  var entryQuantity float64
  if q.Get("entry_price") != "" {
    entryPrice, _ = strconv.ParseFloat(q.Get("entry_price"), 64)
  } else if q.Get("entryPrice") != "" {
    entryPrice, _ = strconv.ParseFloat(q.Get("entryPrice"), 64)
  }

  if q.Get("entry_quantity") != "" {
    entryQuantity, _ = strconv.ParseFloat(q.Get("entry_quantity"), 64)
  } else if q.Get("entryQuantity") != "" {
    entryQuantity, _ = strconv.ParseFloat(q.Get("entryQuantity"), 64)
  }

  if entryPrice == 0 || entryQuantity == 0 {
    position, err := h.PositionsRepository.Get(symbol, side)
    if err == nil && position != nil {
      if entryPrice == 0 {
        entryPrice = position.EntryPrice
      }
      if entryQuantity == 0 {
        entryQuantity = position.EntryQuantity
      }
    }
  }

  if entryPrice <= 0 {
    h.Response.Error(w, http.StatusForbidden, 1004, "entry price is invalid")
    return
  }
  if entryQuantity <= 0 {
    h.Response.Error(w, http.StatusForbidden, 1004, "entry quantity is invalid")
    return
  }

  entity, err := h.SymbolsRepository.Get(symbol)
  if err != nil {
    h.Response.Error(w, http.StatusForbidden, 1004, "symbol not exists")
    return
  }

  tickSize, stepSize, notional, err := h.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    h.Response.Error(w, http.StatusForbidden, 1004, "symbol filters not exists")
    return
  }

  entryPriceDec := decimal.NewFromFloat(entryPrice)
  entryQuantityDec := decimal.NewFromFloat(entryQuantity)
  entryAmountDec := entryPriceDec.Mul(entryQuantityDec)

  stepSizeDec := decimal.NewFromFloat(stepSize)
  notionalDec := decimal.NewFromFloat(notional)

  if stepSize > 0 {
    entryQuantityDec = entryAmountDec.Div(entryPriceDec).Div(stepSizeDec).Floor().Mul(stepSizeDec)
  } else {
    entryQuantityDec = entryAmountDec.Div(entryPriceDec)
  }

  takePrice := h.GamblingRepository.TakePrice(entryPrice, side, tickSize)
  stopPrice := h.GamblingRepository.StopPrice(entryPrice, side, tickSize)
  takePriceDec := decimal.NewFromFloat(takePrice)

  planPrice := entryPrice
  planQuantityDec := entryQuantityDec
  planAmountDec := decimal.Zero
  planProfitDec := decimal.Zero
  lastProfitDec := decimal.Zero

  result := &GamblingCalcResponse{
    TakePrice: takePrice,
    StopPrice: stopPrice,
    Plans:     make([]*GamblingPlanInfo, 0),
  }

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
        h.Response.Error(w, http.StatusForbidden, 1004, "plan amount less than notional")
        return
      }

      takeProfitFloat, _ := takeProfitDec.Float64()
      planAmountFloat, _ := planAmountDec.Float64()
      planProfitFloat, _ := planProfitDec.Float64()

      result.Plans = append(result.Plans, &GamblingPlanInfo{
        Price:      plan.TakePrice,
        Quantity:   plan.TakeQuantity,
        TakeProfit: takeProfitFloat,
        Amount:     planAmountFloat,
        Profit:     planProfitFloat,
      })
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
      h.Response.Error(w, http.StatusForbidden, 1004, "plan amount less than notional")
      return
    }

    planQuantityFloat, _ := planQuantityDec.Float64()
    takeProfitFloat, _ := takeProfitDec.Float64()
    planAmountFloat, _ := planAmountDec.Float64()
    planProfitFloat, _ := planProfitDec.Float64()

    result.Plans = append(result.Plans, &GamblingPlanInfo{
      Price:      takePrice,
      Quantity:   planQuantityFloat,
      TakeProfit: takeProfitFloat,
      Amount:     planAmountFloat,
      Profit:     planProfitFloat,
    })
  }

  result.PlanProfit, _ = planProfitDec.Float64()

  h.Response.Json(w, result)
}
