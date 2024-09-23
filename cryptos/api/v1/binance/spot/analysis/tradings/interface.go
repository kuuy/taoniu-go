package tradings

type ScalpingInfo struct {
  ID             string `json:"id"`
  Day            string `json:"day"`
  BuysCount      int    `json:"buys_count"`
  SellsCount     int    `json:"sells_count"`
  BuysAmount     string `json:"buys_amount"`
  SellsAmount    string `json:"sells_amount"`
  Profit         string `json:"profit"`
  AdditiveProfit string `json:"additive_profit"`
}

type TriggerInfo struct {
  ID             string `json:"id"`
  Day            string `json:"day"`
  BuysCount      int    `json:"buys_count"`
  SellsCount     int    `json:"sells_count"`
  BuysAmount     string `json:"buys_amount"`
  SellsAmount    string `json:"sells_amount"`
  Profit         string `json:"profit"`
  AdditiveProfit string `json:"additive_profit"`
}
