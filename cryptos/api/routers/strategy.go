package routers

import (
  "encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	pool "taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories"
)

type StrategyHandler struct{
  repository *repositories.StrategyRepository
}

type Strategy struct{
  ID string `json:"id"`
  Symbol string `json:"symbol"`
  Indicator string `json:"indicator"`
  Price float64 `json:"price"`
  Signal int64 `json:"signal"`
}

type StrategyDetail struct{
  ID string `json:"id"`
  Symbol string `json:"symbol"`
  Indicator string `json:"indicator"`
  Price float64 `json:"price"`
  Signal int64 `json:"signal"`
}

type ListStrategyResponse struct{
  Strategies []Strategy `json:"strategies"`
}

func NewStrategyRouter() http.Handler {
  db := pool.NewDB()
  repository := repositories.NewStrategyRepository(db)

  handler := StrategyHandler{
    repository : repository,
  }

  r := chi.NewRouter()
  r.Get("/", handler.Listings)

  return r
}

func (h *StrategyHandler) Listings(
  w http.ResponseWriter,
  r *http.Request,
) {
  strategies, err := h.repository.Listings()
  if err != nil {
    return
  }

  var response ListStrategyResponse
  for _,entity := range(strategies) {
    var strategy Strategy
    strategy.ID = entity.ID
    strategy.Symbol = entity.Symbol
    strategy.Indicator = entity.Indicator
    strategy.Price = entity.Price
    strategy.Signal = entity.Signal

    response.Strategies = append(response.Strategies, strategy)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  jsonResponse, err := json.Marshal(response)
  if err != nil {
    return
  }

  w.Write(jsonResponse)
}

