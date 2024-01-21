package tradingview

import (
  "github.com/go-chi/chi/v5"
  "net/http"
  "taoniu.local/cryptos/common"

  "taoniu.local/cryptos/api"
)

type DatafeedHandler struct {
  ApiContext *common.ApiContext
  Response   *api.ResponseHandler
}

func NewDatafeedRouter(apiContext *common.ApiContext) http.Handler {
  h := DatafeedHandler{
    ApiContext: apiContext,
  }

  r := chi.NewRouter()
  r.Get("/config", h.Config)

  return r
}

func (h *DatafeedHandler) Config(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  var config map[string]interface{}
  config["exchanges"] = []map[string]interface{}{}
  config["symbols_types"] = []map[string]interface{}{
    {
      "name":  "Scalping",
      "value": "scalping",
    },
    {
      "name":  "Triggers",
      "value": "triggers",
    },
  }
  config["supported_resolutions"] = []string{
    "1",
    "15",
    "240",
    "D",
    "6M",
  }
  config["supports_search"] = true
  config["supports_group_request"] = true
  config["supports_marks"] = true
  config["supports_timescale_marks"] = true
  config["supports_time"] = true

  h.Response.Out(config)
}
