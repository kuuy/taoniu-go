package spot

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/go-redis/redis/v8"
	"log"
	config "taoniu.local/cryptos/config/binance"
)

type SymbolsRepository struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (r *SymbolsRepository) Flush() error {
	client := binance.NewClient(config.REST_API_KEY, config.REST_SECRET_KEY)
	result, err := client.NewExchangeInfoService().Do(r.Ctx)
	if err != nil {
		return err
	}
	for _, item := range result.Symbols {
		log.Println("symbol:", item.Symbol, item.Status, item.QuotePrecision, item.QuoteAssetPrecision, item.BaseCommissionPrecision, item.QuoteCommissionPrecision)
	}
	return nil
}
