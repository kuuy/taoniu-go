package fishers

import (
	"gorm.io/gorm"

	savingsModels "taoniu.local/cryptos/models/binance/savings"
	models "taoniu.local/cryptos/models/binance/spot/tradings/fishers"
)

type ProductsRepository interface {
	Get(asset string) (savingsModels.FlexibleProduct, error)
	Purchase(productId string, amount float64) (int64, error)
}

type GridsRepository struct {
	Db                 *gorm.DB
	AccountRepository  AccountRepository
	ProductsRepository ProductsRepository
}

type AnalysisInfo struct {
	Symbol   string
	Quantity float64
}

func (r *GridsRepository) Pending() map[string]float64 {
	var result []*AnalysisInfo
	r.Db.Model(&models.Grid{}).Select(
		"symbol",
		"sum(sell_quantity) as quantity",
	).Where("status", 1).Group("symbol").Find(&result)
	data := make(map[string]float64)
	for _, item := range result {
		data[item.Symbol] = item.Quantity
	}
	return data
}
