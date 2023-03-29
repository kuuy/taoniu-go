package tradings

import (
	"gorm.io/gorm"
	fishersRepositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
)

type SymbolsRepository struct {
	Db                *gorm.DB
	FishersRepository *fishersRepositories.FishersRepository
}

func (r *SymbolsRepository) Fishers() *fishersRepositories.FishersRepository {
	if r.FishersRepository == nil {
		r.FishersRepository = &fishersRepositories.FishersRepository{
			Db: r.Db,
		}
	}
	return r.FishersRepository
}

func (r *SymbolsRepository) Scan() []string {
	var symbols []string
	for _, symbol := range r.Fishers().Scan() {
		if !r.contains(symbols, symbol) {
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

func (r *SymbolsRepository) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
