package futures

type OrderInfo struct {
  ID           string  `json:"id"`
  Symbol       string  `json:"symbol"`
  OrderId      int64   `json:"order_id"`
  Type         string  `json:"type"`
  PositionSide string  `json:"position_side"`
  Side         string  `json:"side"`
  Price        float64 `json:"price"`
  Quantity     float64 `json:"quantity"`
  OpenTime     int64   `json:"open_time"`
  UpdateTime   int64   `json:"update_time"`
  ReduceOnly   bool    `json:"reduce_only"`
  Status       string  `json:"status"`
}

type StrategiesInfo struct {
  ID        string  `json:"id"`
  Symbol    string  `json:"symbol"`
  Indicator string  `json:"indicator"`
  Interval  string  `json:"interval"`
  Signal    int     `json:"signal"`
  Price     float64 `json:"price"`
  Timestamp int64   `json:"timestamp"`
}

type PlansInfo struct {
  ID        string  `json:"id"`
  Symbol    string  `json:"symbol"`
  Side      int     `json:"side"`
  Interval  string  `json:"interval"`
  Price     float64 `json:"price"`
  Quantity  float64 `json:"quantity"`
  Amount    float64 `json:"amount"`
  Timestamp int64   `json:"timestamp"`
  Status    int     `json:"status"`
}

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

type SignalInfo struct {
  Price     float64 `json:"price"`
  Signal    int     `json:"signal"`
  Timestamp int64   `json:"timestamp"`
}

type TradingInfo struct {
  BuyPrice      float64 `json:"buy_price"`
  SellPrice     float64 `json:"sell_price"`
  Quantity      float64 `json:"quantity"`
  EntryPrice    float64 `json:"entry_price"`
  EntryQuantity float64 `json:"entry_quantity"`
}

type PositionCalcResponse struct {
  TakePrice float64        `json:"take_price"`
  StopPrice float64        `json:"stop_price"`
  Tradings  []*TradingInfo `json:"tradings"`
}

type GamblingPlanInfo struct {
  Price    float64 `json:"price"`
  Quantity float64 `json:"quantity"`
  Profit   string  `json:"profit"`
}

type CalcGamblingResponse struct {
  TakePrice   float64             `json:"take_price"`
  StopPrice   float64             `json:"stop_price"`
  PlansProfit string              `json:"plans_profit"`
  Plans       []*GamblingPlanInfo `json:"plans"`
}

type ScalpingInfo struct {
  ID          string  `json:"id"`
  Symbol      string  `json:"symbol"`
  Side        int     `json:"side"`
  Capital     float64 `json:"capital"`
  Price       float64 `json:"price"`
  TakePrice   float64 `json:"take_price"`
  StopPrice   float64 `json:"stop_price"`
  TakeOrderId int64   `json:"take_order_id"`
  StopOrderId int64   `json:"stop_order_id"`
  Profit      float64 `json:"profit"`
  Timestamp   int64   `json:"timestamp"`
  Status      int     `json:"status"`
  ExpiredAt   int64   `json:"expired_at"`
  CreatedAt   int64   `json:"created_at"`
}

type TriggersInfo struct {
  ID          string  `json:"id"`
  Symbol      string  `json:"symbol"`
  Side        int     `json:"side"`
  Capital     float64 `json:"capital"`
  Price       float64 `json:"price"`
  TakePrice   float64 `json:"take_price"`
  StopPrice   float64 `json:"stop_price"`
  TakeOrderId int64   `json:"take_order_id"`
  StopOrderId int64   `json:"stop_order_id"`
  Profit      float64 `json:"profit"`
  Timestamp   int64   `json:"timestamp"`
  Status      int     `json:"status"`
  ExpiredAt   int64   `json:"expired_at"`
  CreatedAt   int64   `json:"created_at"`
}
