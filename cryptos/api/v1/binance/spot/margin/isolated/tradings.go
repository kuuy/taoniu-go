package isolated

import (
  "net/http"

  "github.com/go-chi/chi/v5"
  "gorm.io/gorm"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/api/v1/binance/spot/margin/isolated/tradings"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"
  tradingsRepositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type TradingsHandler struct {
  Db         *gorm.DB
  Response   *api.ResponseHandler
  Repository *repositories.TradingsRepository
}

func NewTradingsRouter() http.Handler {
  h := TradingsHandler{
    Db: common.NewDB(),
  }
  h.Repository = &repositories.TradingsRepository{
    Db: h.Db,
  }
  h.Repository.FishersRepository = &tradingsRepositories.FishersRepository{
    Db: h.Db,
  }

  r := chi.NewRouter()
  r.Get("/scan", h.Scan)
  r.Mount("/fishers", tradings.NewFishersRouter())
  return r
}

func (h *TradingsHandler) Scan(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }
  symbols := h.Repository.Scan()
  h.Response.Json(symbols)
}
