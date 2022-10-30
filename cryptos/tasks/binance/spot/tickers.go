package spot

import (
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersTask struct {
	Db         *gorm.DB
	Repository *repositories.TickersRepository
}

func (t *TickersTask) Flush() error {
	var symbols []string
	t.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for i := 0; i < len(symbols); i += 20 {
		j := i + 20
		if j > len(symbols)-1 {
			j = len(symbols) - 1
		}
		t.Repository.Flush(symbols[i:j])
	}

	return nil
}
