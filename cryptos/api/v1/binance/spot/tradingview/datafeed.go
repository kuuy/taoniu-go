package tradingview

import (
  "fmt"
  "math"
  "net/http"
  "strconv"
  "strings"
  "time"

  "github.com/go-chi/chi/v5"

  "taoniu.local/cryptos/api"
  "taoniu.local/cryptos/common"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
)

type DatafeedHandler struct {
  ApiContext         *common.ApiContext
  Response           *api.ResponseHandler
  SymbolsRepository  *spotRepositories.SymbolsRepository
  KlinesRepository   *spotRepositories.KlinesRepository
  TickersRepository  *spotRepositories.TickersRepository
  ScalpingRepository *spotRepositories.ScalpingRepository
  TriggersRepository *spotRepositories.TriggersRepository
}

type SearchInfo struct {
  Symbol       string   `json:"symbol"`
  FullName     string   `json:"full_name"`
  Description  string   `json:"description"`
  Exchange     string   `json:"exchange"`
  Type         string   `json:"type"`
  LogoUrls     []string `json:"logo_urls"`
  ExchangeLogo string   `json:"exchange_logo"`
}

type SymbolInfo struct {
  Symbol               string   `json:"symbol"`
  Name                 string   `json:"name"`
  FullName             string   `json:"full_name"`
  Ticker               string   `json:"ticker"`
  CurrencyCode         string   `json:"currency_code"`
  Exchange             string   `json:"exchange"`
  ExchangeTraded       string   `json:"exchange-traded"`
  ExchangeListed       string   `json:"exchange-listed"`
  Type                 string   `json:"type"`
  Session              string   `json:"session"`
  Timezone             string   `json:"timezone"`
  MinMov               int      `json:"minmov"`
  MinMovV2             int      `json:"minmov2"`
  PriceScale           float64  `json:"pricescale"`
  SupportedResolutions []string `json:"supported_resolutions"`
  HasIntraday          bool     `json:"has_intraday"`
  HasDaily             bool     `json:"has_daily"`
  HasWeeklyAndMonthly  bool     `json:"has_weekly_and_monthly"`
  DataStatus           bool     `json:"data_status"`
  Description          string   `json:"description"`
  LogoUrls             []string `json:"logo_urls"`
  ExchangeLogo         string   `json:"exchange_logo"`
}

type HistoryInfo struct {
  Status    string    `json:"s"`
  Timestamp []int64   `json:"t"`
  Open      []float64 `json:"o"`
  High      []float64 `json:"h"`
  Low       []float64 `json:"l"`
  Close     []float64 `json:"c"`
  Volume    []float64 `json:"v"`
}

func NewDatafeedRouter(apiContext *common.ApiContext) http.Handler {
  h := DatafeedHandler{
    ApiContext: apiContext,
  }
  h.SymbolsRepository = &spotRepositories.SymbolsRepository{
    Db: h.ApiContext.Db,
  }
  h.KlinesRepository = &spotRepositories.KlinesRepository{
    Db: h.ApiContext.Db,
  }
  h.TickersRepository = &spotRepositories.TickersRepository{
    Rdb: h.ApiContext.Rdb,
    Ctx: h.ApiContext.Ctx,
  }
  h.ScalpingRepository = &spotRepositories.ScalpingRepository{
    Db: h.ApiContext.Db,
  }
  h.TriggersRepository = &spotRepositories.TriggersRepository{
    Db: h.ApiContext.Db,
  }

  r := chi.NewRouter()
  r.Get("/time", h.Time)
  r.Get("/config", h.Config)
  r.Get("/search", h.Search)
  r.Get("/symbols", h.SymbolInfo)
  r.Get("/history", h.History)

  return r
}

func (h *DatafeedHandler) Time(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  timestamp := time.Now().Unix()

  h.Response.Out(timestamp)
}

func (h *DatafeedHandler) Config(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  config := map[string]interface{}{}
  config["exchanges"] = []map[string]interface{}{}
  config["symbols_types"] = []map[string]interface{}{
    {
      "name":  "All",
      "value": "",
    },
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
  config["supports_group_request"] = false
  config["supports_marks"] = false
  config["supports_timescale_marks"] = false
  config["supports_time"] = true

  h.Response.Out(config)
}

func (h *DatafeedHandler) Search(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  q := r.URL.Query().Get("query")
  t := r.URL.Query().Get("type")

  var limit int
  if !r.URL.Query().Has("limit") {
    limit = 50
  } else {
    limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
  }

  var symbols []string
  if t == "" {
    symbols = h.SymbolsRepository.Symbols()
  }
  if t == "scalping" {
    symbols = h.ScalpingRepository.Scan()
  }
  if t == "triggers" {
    symbols = h.TriggersRepository.Scan()
  }

  var result []*SearchInfo
  for _, symbol := range symbols {
    if q != "" && !strings.Contains(symbol, q) {
      continue
    }
    result = append(result, &SearchInfo{
      Symbol:      symbol,
      FullName:    symbol,
      Description: symbol,
      Exchange:    "BINANCE",
      Type:        "spot",
    })
    if len(result) == limit {
      break
    }
  }

  h.Response.Out(result)
}

func (h *DatafeedHandler) SymbolInfo(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  symbol := r.URL.Query().Get("symbol")
  entity, err := h.SymbolsRepository.Get(symbol)
  if err != nil {
    h.Response.Out(map[string]interface{}{
      "s":      "error",
      "errmsg": fmt.Sprintf("unknown_symbol %v", symbol),
    })
    return
  }
  tickSize, _, _, err := h.SymbolsRepository.Filters(entity.Filters)
  if err != nil {
    h.Response.Out(map[string]interface{}{
      "s":      "error",
      "errmsg": fmt.Sprintf("symbil filters empty %v", symbol),
    })
    return
  }

  result := &SymbolInfo{
    Symbol:              symbol,
    Name:                symbol,
    FullName:            symbol,
    Ticker:              symbol,
    CurrencyCode:        entity.QuoteAsset,
    Exchange:            "BINANCE",
    ExchangeTraded:      "BINANCE",
    ExchangeListed:      "BINANCE",
    Type:                "spot crypto",
    Session:             "24x7",
    Timezone:            "UTC",
    MinMov:              1,
    MinMovV2:            0,
    PriceScale:          math.Round(1 / tickSize),
    HasIntraday:         true,
    HasDaily:            true,
    HasWeeklyAndMonthly: true,
    SupportedResolutions: []string{
      "1",
      "15",
      "240",
      "D",
      "6M",
    },
  }

  h.Response.Out(result)
}

func (h *DatafeedHandler) History(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.ApiContext.Mux.Lock()
  defer h.ApiContext.Mux.Unlock()

  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  symbol := r.URL.Query().Get("symbol")
  resolution := r.URL.Query().Get("resolution")

  var interval string
  if resolution == "1" {
    interval = "1m"
  } else if resolution == "15" {
    interval = "15m"
  } else if resolution == "240" {
    interval = "4h"
  } else if resolution == "1D" {
    interval = "1d"
  }

  from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
  if err != nil {
    h.Response.Out(map[string]interface{}{
      "s":      "error",
      "errmsg": fmt.Sprintf("invalid request"),
    })
    return
  }

  to, err := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
  if err != nil {
    h.Response.Out(map[string]interface{}{
      "s":      "error",
      "errmsg": fmt.Sprintf("invalid request"),
    })
    return
  }

  limit, err := strconv.Atoi(r.URL.Query().Get("countback"))
  if err != nil {
    h.Response.Out(map[string]interface{}{
      "s":      "error",
      "errmsg": fmt.Sprintf("invalid request"),
    })
    return
  }

  klines := h.KlinesRepository.History(symbol, interval, from*1000, to*1000, limit)
  if len(klines) == 0 {
    h.Response.Out(map[string]interface{}{
      "s":        "no_data",
      "nextTime": 0,
    })
    return
  }

  result := &HistoryInfo{
    Status:    "ok",
    Timestamp: []int64{},
    Open:      []float64{},
    High:      []float64{},
    Low:       []float64{},
    Close:     []float64{},
    Volume:    []float64{},
  }
  for _, kline := range klines {
    result.Timestamp = append([]int64{kline.Timestamp / 1000}, result.Timestamp...)
    result.Open = append([]float64{kline.Open}, result.Open...)
    result.High = append([]float64{kline.High}, result.High...)
    result.Low = append([]float64{kline.Low}, result.Low...)
    result.Close = append([]float64{kline.Close}, result.Close...)
    result.Volume = append([]float64{kline.Volume}, result.Volume...)
  }

  h.Response.Out(result)
}
