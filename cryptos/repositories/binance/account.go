package binance

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"strconv"
)

type AccountRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *AccountRepository) Balances() (map[string]float64, error) {
	balances := make(map[string]float64)

	spotBalance, err := r.SpotBalance()
	if err != nil {
		return nil, err
	}
	if spotBalance > 0 {
		balances["spot"] = spotBalance
	}

	isolatedMarginBalance, err := r.IsolatedMarginBalance()
	if err != nil {
		return nil, err
	}
	if isolatedMarginBalance > 0 {
		balances["isolated_margin"] = isolatedMarginBalance
	}

	futuresBalance, err := r.FuturesBalance()
	if err != nil {
		return nil, err
	}
	if futuresBalance > 0 {
		balances["futures"] = futuresBalance
	}

	return balances, nil
}

func (r *AccountRepository) SpotBalance() (float64, error) {
	var balance float64
	currencies, _ := r.Rdb.SMembers(r.Ctx, "binance:spot:currencies").Result()
	for _, currency := range currencies {
		quantityVal, err := r.Rdb.HGet(
			r.Ctx,
			fmt.Sprintf(
				"binance:spot:balances:%s",
				currency,
			),
			"free",
		).Result()
		if err != nil {
			return 0, err
		}
		quantity, _ := strconv.ParseFloat(quantityVal, 64)
		priceVal, err := r.Rdb.HGet(
			r.Ctx,
			fmt.Sprintf(
				"binance:spot:realtime:%sBUSD",
				currency,
			),
			"price",
		).Result()
		if err != nil {
			log.Println("currency not exists", currency)
			continue
		}
		price, _ := strconv.ParseFloat(priceVal, 64)
		balance += quantity * price
	}
	return balance, nil
}

func (r *AccountRepository) IsolatedMarginBalance() (float64, error) {
	var balance float64
	symbols, _ := r.Rdb.SMembers(r.Ctx, "binance:spot:margin:isolated:symbols").Result()
	for _, symbol := range symbols {
		fields := []string{
			"base_borrowed",
			"base_total_asset",
			"quote_borrowed",
			"quote_total_asset",
		}
		data, _ := r.Rdb.HMGet(
			r.Ctx,
			fmt.Sprintf(
				"binance:spot:margin:isolated:balances:%s",
				symbol,
			),
			fields...,
		).Result()
		priceVal, err := r.Rdb.HGet(
			r.Ctx,
			fmt.Sprintf(
				"binance:spot:realtime:%s",
				symbol,
			),
			"price",
		).Result()
		if err != nil {
			return 0, err
		}
		price, _ := strconv.ParseFloat(priceVal, 64)
		baseBorrowed, _ := strconv.ParseFloat(data[0].(string), 64)
		baseTotalAsset, _ := strconv.ParseFloat(data[1].(string), 64)
		quoteBorrowed, _ := strconv.ParseFloat(data[2].(string), 64)
		quoteTotalAsset, _ := strconv.ParseFloat(data[3].(string), 64)
		balance += (baseTotalAsset - baseBorrowed) * price
		balance -= quoteBorrowed
		balance += quoteTotalAsset
	}

	return balance, nil
}

func (r *AccountRepository) FuturesBalance() (float64, error) {
	var balance float64
	return balance, nil
}
