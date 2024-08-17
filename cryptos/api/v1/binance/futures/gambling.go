package futures

import (
  "fmt"
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  "github.com/shopspring/decimal"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
)

type GamblingHandler struct {
  ApiContext          *common.ApiContext
  Response            *api.ResponseHandler
  Repository          *futuresRepositories.GamblingRepository
  PositionsRepository *futuresRepositories.PositionsRepository
}

func NewGamblingRouter(apiContext *common.ApiContext) http.Handler {
  h := GamblingHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.Repository = &futuresRepositories.GamblingRepository{
    Db: h.ApiContext.Db,
  }
  h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
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
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

  q := r.URL.Query()

  if q.Get("symbol") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "symbol is empty")
    return
  }

  if q.Get("side") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "side is empty")
    return
  }

  if q.Get("entry_price") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "entry price is empty")
    return
  }

  if q.Get("entry_quantity") == "" {
    h.Response.Error(http.StatusForbidden, 1004, "entry quantity is empty")
    return
  }

  symbol := q.Get("symbol")
  side, _ := strconv.Atoi(q.Get("side"))
  entryPrice, _ := strconv.ParseFloat(q.Get("entry_price"), 16)
  entryQuantity, _ := strconv.ParseFloat(q.Get("entry_quantity"), 16)
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

  tickSize, stepSize, err := h.Repository.Filters(symbol)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1004, "symbol filters not exists")
    return
  }

  takePrice := h.Repository.TakePrice(entryPrice, side, tickSize)
  stopPrice := h.Repository.StopPrice(entryPrice, side, tickSize)

  result := &CalcGamblingResponse{}

  planPrice := entryPrice
  planQuantity := entryQuantity
  planAmount := entryAmount
  planProfit := 0.0
  lastProfit := 0.0
  takeProfit := 0.0

  for {
    plans := h.Repository.Calc(planPrice, planQuantity, side, tickSize, stepSize)
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
      result.Plans = append(result.Plans, &GamblingPlanInfo{
        Price:    plan.TakePrice,
        Quantity: plan.TakeQuantity,
        Profit:   fmt.Sprintf("%.2f", takeProfit),
      })
    }
    if len(plans) == 0 || lastProfit > 0 {
      break
    }
  }

  planProfit, _ = decimal.NewFromFloat(planProfit).Add(decimal.NewFromFloat(lastProfit)).Float64()

  if planQuantity > 0 {
    if side == 1 {
      takeProfit, _ = decimal.NewFromFloat(takePrice).Sub(decimal.NewFromFloat(entryPrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
    } else {
      takeProfit, _ = decimal.NewFromFloat(entryPrice).Sub(decimal.NewFromFloat(takePrice)).Mul(decimal.NewFromFloat(planQuantity)).Float64()
    }
    result.Plans = append(result.Plans, &GamblingPlanInfo{
      Price:    takePrice,
      Quantity: planQuantity,
      Profit:   fmt.Sprintf("%.2f", takeProfit),
    })
  }

  result.TakePrice = takePrice
  result.StopPrice = stopPrice
  result.PlansProfit = fmt.Sprintf("%.2f", planProfit)

  h.Response.Json(result)
}
