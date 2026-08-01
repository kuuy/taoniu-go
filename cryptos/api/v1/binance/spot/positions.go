package spot

import (
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"taoniu.local/cryptos/api"
	"taoniu.local/cryptos/common"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type PositionsHandler struct {
  ApiContext        *common.ApiContext
  Response          *api.ResponseHandler
  Repository        *repositories.PositionsRepository
  SymbolsRepository *repositories.SymbolsRepository
}

func NewPositionsRouter(apiContext *common.ApiContext) http.Handler {
  h := PositionsHandler{
    ApiContext: apiContext,
  }
  h.Response = &api.ResponseHandler{}
  h.Response.Jwe = &common.Jwe{}
  h.Repository = &repositories.PositionsRepository{
    Db: h.ApiContext.Db,
  }
  h.Repository.SymbolsRepository = &repositories.SymbolsRepository{
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
  conditions := make(map[string]interface{})

  positions := h.Repository.Gets(conditions)
  data := make([]*PositionsInfo, len(positions))
  for i, position := range positions {
    data[i] = &PositionsInfo{
      ID:            position.ID,
      Symbol:        position.Symbol,
      Notional:      position.Notional,
      EntryPrice:    position.EntryPrice,
      EntryQuantity: position.EntryQuantity,
      EntryAmount:   position.EntryPrice * position.EntryQuantity,
      Timestamp:     position.Timestamp,
    }
  }

  h.Response.Json(w, data)
}

func (h *PositionsHandler) Calc(
	w http.ResponseWriter,
	r *http.Request,
) {
	q := r.URL.Query()

	if q.Get("symbol") == "" {
		h.Response.Error(w, http.StatusForbidden, 1004, "symbol is empty")
		return
	}

	if q.Get("max_capital") == "" {
		h.Response.Error(w, http.StatusForbidden, 1004, "max_capital is empty")
		return
	}

	if q.Get("entry_price") == "" {
		h.Response.Error(w, http.StatusForbidden, 1004, "entry_price is empty")
		return
	}

	if q.Get("entry_quantity") == "" {
		h.Response.Error(w, http.StatusForbidden, 1004, "entry_quantity is empty")
		return
	}

	symbol := q.Get("symbol")
	maxCapital, _ := strconv.ParseFloat(q.Get("max_capital"), 64)

	entryPrice, _ := strconv.ParseFloat(q.Get("entry_price"), 64)
	entryQuantity, _ := strconv.ParseFloat(q.Get("entry_quantity"), 64)
	entryAmount, _ := decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

	tickSize, stepSize, err := h.Repository.Filters(symbol)
	if err != nil {
		h.Response.Error(w, http.StatusForbidden, 1004, "symbol filters not exists")
		return
	}

	if stepSize > 0 {
		entryQuantity, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryPrice)).Div(decimal.NewFromFloat(stepSize)).Floor().Mul(decimal.NewFromFloat(stepSize)).Float64()
	} else {
		entryQuantity, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryPrice)).Float64()
	}
	entryAmount, _ = decimal.NewFromFloat(entryPrice).Mul(decimal.NewFromFloat(entryQuantity)).Float64()

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
		sellPrice = h.Repository.SellPrice(entryPrice, entryAmount)
		sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
		takePrice = h.Repository.TakePrice(entryPrice, tickSize)
	} else {
		takePrice = h.Repository.TakePrice(entryPrice, tickSize)
	}

	ipart, _ := math.Modf(maxCapital)
	places := 1
	for ; ipart >= 10; ipart = ipart / 10 {
		places++
	}

	result := &PositionCalcResponse{}
	priceRatio := 1.0

	for {
		if buyPrice > 0 && entryPrice > 0 {
			priceRatio, _ = decimal.NewFromFloat(entryPrice).Div(decimal.NewFromFloat(buyPrice)).Float64()
		}

		var err error
		capital, err := h.Repository.Capital(maxCapital, entryAmount, places, priceRatio)
		if err != nil {
			break
		}
		ratio := h.Repository.Ratio(capital, entryAmount, priceRatio)
		buyAmount, _ = decimal.NewFromFloat(capital).Mul(decimal.NewFromFloat(ratio)).Float64()
		if buyAmount < 5 {
			buyAmount = 5
		}
		buyQuantity = h.Repository.BuyQuantity(buyAmount, entryPrice, entryAmount)
		buyPrice, _ = decimal.NewFromFloat(buyAmount).Div(decimal.NewFromFloat(buyQuantity)).Float64()
		buyPrice, _ = decimal.NewFromFloat(buyPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()
		buyQuantity, _ = decimal.NewFromFloat(buyQuantity).Div(decimal.NewFromFloat(stepSize)).Ceil().Mul(decimal.NewFromFloat(stepSize)).Float64()
		buyAmount, _ = decimal.NewFromFloat(buyPrice).Mul(decimal.NewFromFloat(buyQuantity)).Float64()
		entryQuantity, _ = decimal.NewFromFloat(entryQuantity).Add(decimal.NewFromFloat(buyQuantity)).Float64()
		entryAmount, _ = decimal.NewFromFloat(entryAmount).Add(decimal.NewFromFloat(buyAmount)).Float64()
		entryPrice, _ = decimal.NewFromFloat(entryAmount).Div(decimal.NewFromFloat(entryQuantity)).Float64()
		sellPrice = h.Repository.SellPrice(entryPrice, entryAmount)
		sellPrice, _ = decimal.NewFromFloat(sellPrice).Div(decimal.NewFromFloat(tickSize)).Ceil().Mul(decimal.NewFromFloat(tickSize)).Float64()
		result.Tradings = append(result.Tradings, &TradingInfo{
			BuyPrice:      buyPrice,
			SellPrice:     sellPrice,
			Quantity:      buyQuantity,
			EntryPrice:    entryPrice,
			EntryQuantity: entryQuantity,
		})
	}

	stopAmount, _ := decimal.NewFromFloat(entryAmount).Mul(decimal.NewFromFloat(0.1)).Float64()

	var stopPrice float64
	stopPrice, _ = decimal.NewFromFloat(entryPrice).Sub(
		decimal.NewFromFloat(stopAmount).Div(decimal.NewFromFloat(entryQuantity)),
	).Float64()
	stopPrice, _ = decimal.NewFromFloat(stopPrice).Div(decimal.NewFromFloat(tickSize)).Floor().Mul(decimal.NewFromFloat(tickSize)).Float64()

	result.TakePrice = takePrice
	result.StopPrice = stopPrice

	h.Response.Json(w, result)
}
