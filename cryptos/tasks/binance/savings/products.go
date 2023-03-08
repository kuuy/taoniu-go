package savings

import repositories "taoniu.local/cryptos/repositories/binance/savings"

type ProductsTask struct {
	Repository *repositories.ProductsRepository
}

func (t *ProductsTask) Flush() error {
	return t.Repository.Flush()
}
