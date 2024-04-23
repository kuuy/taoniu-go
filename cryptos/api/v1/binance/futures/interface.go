package futures

type PositionInfo struct {
  ID            string  `json:"id"`
  Symbol        string  `json:"symbol"`
  Side          int     `json:"side"`
  Leverage      int     `json:"leverage"`
  Capital       float64 `json:"capital"`
  Notional      float64 `json:"notional"`
  EntryPrice    float64 `json:"entry_price"`
  EntryQuantity float64 `json:"entry_quantity"`
  EntryAmount   float64 `json:"entry_amount"`
  Timestamp     int64   `json:"timestamp"`
}

type TradingInfo struct {
  BuyPrice      float64 `json:"buy_price"`
  SellPrice     float64 `json:"sell_price"`
  Quantity      float64 `json:"quantity"`
  EntryPrice    float64 `json:"entry_price"`
  EntryQuantity float64 `json:"entry_quantity"`
}

type CalcPositionResponse struct {
  TakePrice float64        `json:"take_price"`
  StopPrice float64        `json:"stop_price"`
  Tradings  []*TradingInfo `json:"tradings"`
}

type GamblingPlanInfo struct {
  Price    float64 `json:"price"`
  Quantity float64 `json:"quantity"`
  Profit   string  `json:"profit"`
}

type CalcGameblingResponse struct {
  TakePrice   float64             `json:"take_price"`
  StopPrice   float64             `json:"stop_price"`
  PlansProfit string              `json:"plans_profit"`
  Plans       []*GamblingPlanInfo `json:"plans"`
}
