package savings

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"log"
	config "taoniu.local/cryptos/config/binance/spot"
)

type ProductsRepository struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (r *ProductsRepository) Flush() error {
	client := binance.NewClient(config.ACCOUNT_API_KEY, config.ACCOUNT_SECRET_KEY)
	products, err := client.NewListSavingsFixedAndActivityProductsService().Current(1).Size(10).Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, product := range products {
		log.Println("product:", product)
		//	free, _ := strconv.ParseFloat(coin.Free, 64)
		//	if free <= 0.0 {
		//		r.Rdb.Del(r.Ctx, fmt.Sprintf("binance:spot:balances:%s", coin.Asset))
		//		continue
		//	}
		//	r.Rdb.HMSet(
		//		r.Ctx,
		//		fmt.Sprintf("binance:spot:balances:%s", coin.Asset),
		//		map[string]interface{}{
		//			"free":   coin.Free,
		//			"locked": coin.Locked,
		//		},
		//	)
	}

	return nil
}
