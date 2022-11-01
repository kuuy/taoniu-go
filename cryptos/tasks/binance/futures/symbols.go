package futures

import repositories "taoniu.local/cryptos/repositories/binance/futures"

type SymbolsTask struct {
	Repository *repositories.SymbolsRepository
}

func (t *SymbolsTask) Flush() error {
	return t.Repository.Flush()
}

func (t *SymbolsTask) Count() error {
	return t.Repository.Count()
}