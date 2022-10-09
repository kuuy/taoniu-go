package isolated

import repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated"

type SymbolsTask struct {
	Repository *repositories.SymbolsRepository
}

func (t *SymbolsTask) Flush() error {
	return t.Repository.Flush()
}
