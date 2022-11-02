package futures

import (
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance/futures"
	repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type TickersTask struct {
	Db         *gorm.DB
	Repository *repositories.TickersRepository
}

func (t *TickersTask) Flush() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status", "TRADING").Find(&symbols)
	for i := 0; i < len(symbols); i += 20 {
		j := i + 20
		if j > len(symbols)-1 {
			j = len(symbols) - 1
		}
		t.Repository.Flush(symbols[i:j])
	}

	return nil
}
