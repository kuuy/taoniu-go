package futures

import (
  "math"
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  "github.com/shopspring/decimal"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
)

type PositionsHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
  Repository *futuresRepositories.PositionsRepository
}

func NewPositionsRouter(apiContext *common.ApiContext) http.Handler {
  h := PositionsHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.JweRepository = &repositories.JweRepository{}
  h.Repository = &futuresRepositories.PositionsRepository{
    Db: h.ApiContext.Db,
  }
  h.Repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/", h.Gets)
  r.Get("/calc", h.Calc)

  return r
}

func (h *PositionsHandler) Gets(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response.Writer = w

  q := r.URL.Query()
  conditions := make(map[string]interface{})
  if q.Get("side") != "" {
    conditions["side"], _ = strconv.Atoi(q.Get("side"))
  }

  positions := h.Repository.Gets(conditions)
  data := make([]*PositionInfo, len(positions))
  for i, position := range positions {
    data[i] = &PositionInfo{
      ID:            position.ID,
      Symbol:        position.Symbol,
      Side:          position.Side,
      Leverage:      position.Leverage,
      Capital:       position.Capital,
      Notional:      position.Notional,
      EntryPrice:    position.EntryPrice,
      EntryQuantity: position.EntryQuantity,
      EntryAmount:   position.EntryPrice * position.EntryQuantity,
      Timestamp:     position.Timestamp,
    }
  }

  h.Response.Json(data)
}

func (h *PositionsHandler) Calc(
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

  symbol := q.Get("symbol")
  side, _ := strconv.Atoi(q.Get("side"))
  position, err := h.Repository.Get(symbol, side)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1004, "position not exists")
    return
  }

  tickSize, stepSize, err := h.Repository.Filters(symbol)
  if err != nil {
    h.Response.Error(http.StatusForbidden, 1004, "symbol filters not exists")
    return
  }

  entryPrice := position.EntryPrice
  entryQuantity := position.EntryQuantity
  entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

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

  ipart, _ := math.Modf(position.Capital)
  places := 1
  for ; ipart >= 10; ipart = ipart / 10 {
    places++
  }

  result := &PositionCalcResponse{}

  for {
    var err error
    capital, err := h.Repository.Capital(position.Capital, entryAmount, places)
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
    result.Tradings = append(result.Tradings, &TradingInfo{
      BuyPrice:      buyPrice,
      SellPrice:     sellPrice,
      Quantity:      buyQuantity,
      EntryPrice:    entryPrice,
      EntryQuantity: entryQuantity,
    })
  }

  stopAmount, _ := decimal.NewFromFloat(entryAmount).Div(decimal.NewFromInt32(int32(position.Leverage))).Mul(decimal.NewFromFloat(0.1)).Float64()

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

  result.TakePrice = takePrice
  result.StopPrice = stopPrice

  h.Response.Json(result)
}
