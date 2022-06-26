package routers

import (
  "encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	pool "taoniu.local/cryptos/common"
	"taoniu.local/cryptos/repositories"
)

type OrderHandler struct{
  repository *repositories.OrderRepository
}

type Order struct{
  ID string `json:"id"`
  Symbol string `json:"symbol"`
  Type string `json:"type"`
  PositionSide string `json:"position_side"`
  Side string `json:"side"`
  Price float64 `json:"price"`
  Status string `json:"status"`
}

type OrderDetail struct{
  ID string `json:"id"`
  Symbol string `json:"symbol"`
  Type string `json:"type"`
  PositionSide string `json:"position_side"`
  Side string `json:"side"`
  Price float64 `json:"price"`
  Status string `json:"status"`
}

type ListOrderResponse struct{
  Orders []Order `json:"orders"`
}

func NewOrderRouter() http.Handler {
  db := pool.NewDB()
  repository := repositories.NewOrderRepository(db)

  handler := OrderHandler{
    repository : repository,
  }

  r := chi.NewRouter()
  r.Get("/", handler.Listings)

  return r
}

func (h *OrderHandler) Listings(w http.ResponseWriter, r *http.Request) {
  orders, err := h.repository.Listings()
  if err != nil {
    return
  }

  var response ListOrderResponse
  for _,entity := range(orders) {
    var order Order
    order.ID = entity.ID
    order.Symbol = entity.Symbol
    order.Type = entity.Type
    order.PositionSide = entity.PositionSide
    order.Side = entity.Side
    order.Price = entity.Price
    order.Status = entity.Status

    response.Orders = append(response.Orders, order)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  jsonResponse, err := json.Marshal(response)
  if err != nil {
    return
  }

  w.Write(jsonResponse)
}

