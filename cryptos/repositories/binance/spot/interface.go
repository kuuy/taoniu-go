package spot

import savingsModels "taoniu.local/cryptos/models/binance/savings"

type ProductsRepository interface {
  Get(asset string) (savingsModels.FlexibleProduct, error)
  Purchase(productId string, amount float64) (int64, error)
}

type TradingsTriggersRepository interface {
  Scan() []string
}

type RankingResult struct {
  Total int
  Data  []string
}
