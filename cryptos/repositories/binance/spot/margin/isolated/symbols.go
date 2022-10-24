package isolated

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	models "taoniu.local/cryptos/models/binance"
)

type SymbolsRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *SymbolsRepository) Flush() error {
	var symbols []string
	r.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	temp, _ := r.Rdb.SMembers(r.Ctx, "binance:spot:margin:isolated:symbols").Result()
	var margins []string
	for _, symbol := range symbols {
		exists, _ := r.Rdb.Exists(
			r.Ctx,
			fmt.Sprintf("binance:spot:margin:isolated:balances:%s", symbol),
		).Result()
		if exists == 0 {
			continue
		}
		if !r.contains(temp, symbol) {
			r.Rdb.SAdd(
				r.Ctx,
				"binance:spot:margin:isolated:symbols",
				symbol,
			).Result()
		}
		margins = append(margins, symbol)
	}
	for _, symbol := range temp {
		if !r.contains(margins, symbol) {
			r.Rdb.SRem(
				r.Ctx,
				"binance:spot:margin:isolated:symbols",
				symbol,
			)
		}
	}
	return nil
}

func (r *SymbolsRepository) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
