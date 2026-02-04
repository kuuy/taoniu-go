package swap

type ApiResponse struct {
  Success bool `json:"success"`
}

type MintInfo struct {
  Name     string   `json:"name"`
  Symbol   string   `json:"symbol"`
  Address  string   `json:"address"`
  Decimals int      `json:"decimals"`
  Tags     []string `json:"tags"`
}

type MintsListingsResponse struct {
  ApiResponse
  Data struct {
    MintList  []MintInfo `json:"mintList"`
    WhiteList []string   `json:"whiteList"`
  } `json:"data"`
}

type KlineInfo struct {
  Open      float64 `json:"o"`
  Close     float64 `json:"c"`
  High      float64 `json:"h"`
  Low       float64 `json:"l"`
  Volume    float64 `json:"vBase"`
  Quota     float64 `json:"vQuote"`
  Timestamp int64   `json:"unixTime"`
}

type KlinesListingsResponse struct {
  ApiResponse
  Data struct {
    Items []*KlineInfo `json:"items"`
  } `json:"data"`
}
type TickersListingsResponse struct {
  ApiResponse
  Data struct {
    Data []TickerInfo `json:"data"`
  } `json:"data"`
}

type TickerInfo struct {
  PoolId       string  `json:"id"`
  Symbol       string  `json:"symbol"`
  Price        float64 `json:"price"`
  Volume24h    float64 `json:"volume24h"`
  PriceChange  float64 `json:"priceChange24h"`
  LastUpdateAt int64   `json:"lastUpdateAt"`
}
