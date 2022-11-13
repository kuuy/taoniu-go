package dice

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"

	models "taoniu.local/gamblings/models/wolf/dice"
	repositories "taoniu.local/gamblings/repositories/wolf"
)

type MultipleRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	UseProxy          bool
	BetRepository     *BetRepository
	HuntRepository    *HuntRepository
	AccountRepository *repositories.AccountRepository
}

func (r *MultipleRepository) Bet() *BetRepository {
	if r.BetRepository == nil {
		r.BetRepository = &BetRepository{
			Rdb:      r.Rdb,
			Ctx:      r.Ctx,
			UseProxy: r.UseProxy,
		}
	}
	return r.BetRepository
}

func (r *MultipleRepository) Hunt() *HuntRepository {
	if r.HuntRepository == nil {
		r.HuntRepository = &HuntRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.HuntRepository
}

func (r *MultipleRepository) Account() *repositories.AccountRepository {
	if r.AccountRepository == nil {
		r.AccountRepository = &repositories.AccountRepository{
			Rdb:      r.Rdb,
			Ctx:      r.Ctx,
			UseProxy: r.UseProxy,
		}
	}
	return r.AccountRepository
}

func (r *MultipleRepository) Start() error {
	var multiple models.Multiple
	result := r.Db.Where("status", 4).Take(&multiple)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("stop loss occured")
	}

	timestamp := time.Now().Unix()
	r.Rdb.ZAdd(r.Ctx, "wolf:multiple", &redis.Z{
		float64(timestamp),
		"dice",
	})
	return nil
}

func (r *MultipleRepository) Stop() error {
	r.Rdb.ZRem(r.Ctx, "wolf:multiple", "dice")
	return nil
}

func (r *MultipleRepository) Place(currency string) error {
	balance, err := r.Account().Balance(currency)
	if err != nil {
		return err
	}

	amount := math.Ceil(balance*0.000001/0.00000001) / math.Ceil(1/0.00000001)
	stopLoss := math.Ceil(balance*0.3/0.00000001) / math.Ceil(1/0.00000001)

	var multiple models.Multiple
	result := r.Db.Where("status", 3).Take(&multiple)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		multiple = models.Multiple{
			ID:       xid.New().String(),
			Currency: currency,
			Amount:   amount,
			Balance:  balance,
			Invest:   amount,
			Status:   1,
		}
		r.Db.Create(&multiple)
	} else {
		if time.Now().Unix()-multiple.UpdatedAt.Unix() < 5 {
			return errors.New("multiple error waiting")
		}

		if balance > multiple.Balance {
			multiple.WinAmount = math.Ceil((multiple.WinAmount+balance-multiple.Balance)/0.0000000001) / math.Ceil(1/0.0000000001)
			multiple.WinCount += 1
		}
		if balance < multiple.Balance {
			multiple.LossAmount = math.Ceil((multiple.LossAmount+multiple.Balance-balance)/0.0000000001) / math.Ceil(1/0.0000000001)
			multiple.LossCount += 1
		}

		multiple.Profit = math.Ceil((multiple.WinAmount-multiple.LossAmount)/0.0000000001) / math.Ceil(1/0.0000000001)
		if multiple.Profit >= 0 {
			multiple.Status = 2
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Updates(multiple)
			return nil
		}

		if multiple.Invest < -multiple.Profit {
			multiple.Invest = -multiple.Profit
		}

		multiple.Balance = balance
		multiple.Status = 1
		r.Db.Model(&models.Multiple{ID: multiple.ID}).Updates(multiple)

		if (multiple.WinCount > 0 && multiple.WinCount%3 == 0) ||
			(multiple.LossCount > 0 && multiple.LossCount%3 == 0) {
			amount = math.Ceil(-multiple.Profit/(3*0.99*(1.4778-1))/0.00000001) / math.Ceil(1/0.00000001)
		} else {
			amount = multiple.Amount
		}
	}

	rules := []string{"under", "over"}
	for {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(rules), func(i, j int) { rules[i], rules[j] = rules[j], rules[i] })
		rule := rules[(rand.Intn(571-23)+23)%len(rules)]
		var betValue float64
		if rule == "under" {
			betValue = 67
		} else {
			betValue = 33
		}

		hash, result, state, err := r.Bet().Place(currency, amount, rule, betValue)
		if err != nil {
			multiple.Amount = amount
			multiple.Remark = err.Error()
			multiple.Status = 3
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Updates(multiple)
			return err
		}
		r.Hunt().Handing(hash, result)

		multiplier, _ := r.Bet().BetRule(rule, betValue)

		if state {
			multiple.WinAmount = math.Ceil((multiple.WinAmount+amount*(multiplier-1))/0.0000000001) / math.Ceil(1/0.0000000001)
			multiple.WinCount++
		} else {
			multiple.LossAmount = math.Ceil((multiple.LossAmount+amount)/0.00000001) / math.Ceil(1/0.00000001)
			multiple.LossCount++
		}

		multiple.Profit = math.Ceil((multiple.WinAmount-multiple.LossAmount)/0.0000000001) / math.Ceil(1/0.0000000001)
		if multiple.Profit >= 0 {
			break
		}

		if multiple.Invest < -multiple.Profit {
			multiple.Invest = -multiple.Profit
		}

		if -multiple.Profit > stopLoss {
			r.Stop()
			multiple.Status = 4
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Updates(multiple)
			return errors.New("stop loss")
		}

		if (multiple.WinCount > 0 && multiple.WinCount%3 == 0) ||
			(multiple.LossCount > 0 && multiple.LossCount%3 == 0) {
			amount = math.Ceil(-multiple.Profit/(3*0.99*(1.4778-1))/0.00000001) / math.Ceil(1/0.00000001)
		}
	}

	multiple.Status = 2
	r.Db.Model(&models.Multiple{ID: multiple.ID}).Updates(multiple)

	return nil
}

func (r *MultipleRepository) Clean() error {
	expire := time.Now().AddDate(0, 0, -7)
	r.Db.Where("updated_at < ?", expire).Delete(&models.Multiple{})
	return nil
}
