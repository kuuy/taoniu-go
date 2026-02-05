package futures

type AccountInfo struct {
  Assets    []*AssetInfo    `json:"assets"`
  Positions []*PositionInfo `json:"positions"`
}

type AssetInfo struct {
  Asset            string `json:"asset"`
  Balance          string `json:"walletBalance"`
  Free             string `json:"availableBalance"`
  UnrealizedProfit string `json:"unrealizedProfit"`
  Margin           string `json:"marginBalance"`
  InitialMargin    string `json:"initialMargin"`
  MaintMargin      string `json:"maintMargin"`
}

type TickerInfo struct {
  Symbol    string  `json:"symbol"`
  Open      float64 `json:"openPrice,string"`
  Price     float64 `json:"lastPrice,string"`
  Change    float64 `json:"-"`
  High      float64 `json:"highPrice,string"`
  Low       float64 `json:"lowPrice,string"`
  Volume    float64 `json:"volume,string"`
  Quota     float64 `json:"quoteVolume,string"`
  CloseTime int64   `json:"closeTime"`
}

type PositionInfo struct {
  Symbol        string `json:"symbol"`
  PositionSide  string `json:"positionSide"`
  Isolated      bool   `json:"isolated"`
  Leverage      string `json:"leverage"`
  Capital       string `json:"maxNotional"`
  Notional      string `json:"notional"`
  EntryPrice    string `json:"entryPrice"`
  EntryQuantity string `json:"positionAmt"`
  UpdateTime    int64  `json:"updateTime"`
}

type RankingResult struct {
  Total int
  Data  []string
}

type StopLossInfo struct {
  Symbol        string  `json:"symbol"`
  Side          int     `json:"side"`            // 1=LONG, 2=SHORT
  EntryPrice    float64 `json:"entry_price"`     // 入场价
  CurrentPrice  float64 `json:"current_price"`   // 当前价
  InitialStop   float64 `json:"initial_stop"`    // 初始止损
  TrailingStop  float64 `json:"trailing_stop"`   // 移动止损
  BreakEvenStop float64 `json:"break_even_stop"` // 保本止损
  ActiveStop    float64 `json:"active_stop"`     // 当前生效的止损
  TakeProfit1   float64 `json:"take_profit_1"`   // 第一止盈目标
  TakeProfit2   float64 `json:"take_profit_2"`   // 第二止盈目标
  RiskReward    float64 `json:"risk_reward"`     // 风险回报比
  ATR           float64 `json:"atr"`             // ATR值
  ATRMultiplier float64 `json:"atr_multiplier"`  // ATR倍数
  ProfitATR     float64 `json:"profit_atr"`      // 盈利ATR倍数
  Leverage      int     `json:"leverage"`        // 杠杆
  Risk          float64 `json:"risk"`            // 风险比例
  StopType      string  `json:"stop_type"`       // 止损类型
  ShouldTrade   bool    `json:"should_trade"`    // 是否应该交易
  RejectReason  string  `json:"reject_reason"`   // 拒绝原因
}

type GamblingPlan struct {
  TakePrice    float64
  TakeQuantity float64
  TakeAmount   float64
}
