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

type GamblingPlan struct {
  TakePrice    float64
  TakeQuantity float64
  TakeAmount   float64
}
