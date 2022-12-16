package dice

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"taoniu.local/gamblings/common"
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

func (r *MultipleRepository) Apply(currency string) error {
	score, _ := r.Rdb.ZScore(
		r.Ctx,
		"wolf:bet",
		"dice",
	).Result()
	if int64(score) == 0 {
		return errors.New("bet not started")
	}

	var multiple models.Multiple
	result := r.Db.Where("status IN ?", []int{1, 3, 4}).Take(&multiple)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result := r.Db.Order("updated_at desc").Take(&multiple)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if time.Now().Unix()-multiple.UpdatedAt.Unix() < 55 {
				return errors.New("multiple next waiting")
			}
		}

		mutex := common.NewMutex(
			r.Rdb,
			r.Ctx,
			"locks:wolf:dice:apply",
		)
		if mutex.Lock(2 * time.Second) {
			return errors.New("wolf dice apply locked")
		}
		defer mutex.Unlock()

		score, _ := r.Rdb.ZScore(
			r.Ctx,
			"wolf:strategies",
			"dice",
		).Result()
		if int64(score) != 0 {
			return errors.New("bet palace waiting")
		}

		balance, err := r.Account().Balance(currency)
		if err != nil {
			return err
		}

		if balance < 1 {
			return errors.New("balance less than 1")
		}

		amount := math.Round(balance*100000000*0.00000006) / 100000000

		rules := []string{"under", "over"}
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(rules), func(i, j int) { rules[i], rules[j] = rules[j], rules[i] })

		rule := rules[(rand.Intn(313-13)+13)%len(rules)]
		targetBalance := math.Round(balance*10000000000*1.000008) / 10000000000
		stopBalance := math.Round(balance*10000000000*0.95) / 10000000000

		//if stopBalance < balance-10 {
		//	amount = math.Round(20000000000*0.00000006) / 100000000
		//	targetBalance = math.Round(balance*10000000000+2000000000000*0.000008) / 10000000000
		//	stopBalance = math.Round(balance*10000000000-100000000000) / 10000000000
		//}

		multiple = models.Multiple{
			ID:            xid.New().String(),
			Currency:      currency,
			Rule:          rule,
			Amount:        amount,
			Balance:       balance,
			Invest:        amount,
			StartAmount:   amount,
			StartBalance:  balance,
			TargetBalance: targetBalance,
			StopBalance:   stopBalance,
			Status:        1,
		}
		r.Db.Create(&multiple)
	} else {
		if multiple.Status == 4 {
			return errors.New("stop loss occured")
		}
		if multiple.Status == 3 {
			return errors.New("multiple error waiting")
		}
		return errors.New("multiple not finished")
	}

	r.Start()

	return nil
}

func (r *MultipleRepository) Start() {
	r.Rdb.ZAdd(r.Ctx, "wolf:strategies", &redis.Z{
		Score:  3,
		Member: "dice",
	})
}

func (r *MultipleRepository) Stop() {
	r.Rdb.ZRem(r.Ctx, "wolf:strategies", "dice")
}

func (r *MultipleRepository) Switch(betCount int) string {
	rules := []string{"under", "over"}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(rules), func(i, j int) { rules[i], rules[j] = rules[j], rules[i] })
	rule := rules[(rand.Intn(313-13)+13)%len(rules)]

	primes := []int{2, 3, 5, 7, 11}
	for _, prime := range primes {
		if betCount%prime == 0 {
			if rule == "under" {
				rule = "over"
			} else {
				rule = "under"
			}
		}
	}

	return rule
}

func (r *MultipleRepository) Show() (models.Multiple, error) {
	var multiple models.Multiple
	result := r.Db.Where("status IN ?", []int{1, 3}).Take(&multiple)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return multiple, errors.New("multiple empty")
	}
	return multiple, nil
}

func (r *MultipleRepository) Place() error {
	var multiple models.Multiple
	result := r.Db.Where("status IN ?", []int{1, 3}).Take(&multiple)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("multiple empty")
	}

	multiplier := 1.4851
	intervals := []int{1, 2, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40}

	if multiple.Status == 3 {
		if time.Now().Unix()-multiple.UpdatedAt.Unix() < 5 {
			return errors.New("multiple error waiting")
		}

		balance, err := r.Account().Balance(multiple.Currency)
		if err != nil {
			return err
		}

		balanceChange := balance - multiple.Balance

		multiple.Balance = balance
		multiple.Profit = math.Round((balance-multiple.StartBalance)*10000000000) / 10000000000
		multiple.BufferProfit = math.Round((multiple.BufferProfit+balanceChange)*10000000000) / 10000000000

		if multiple.BufferProfit >= 0 {
			multiple.Amount = multiple.StartAmount
			multiple.BufferProfit = 0
		} else {
			ratio := math.Round(multiple.BufferProfit / -multiple.StartAmount)
			if ratio > 0 && int(ratio)%7 == 0 || multiple.StreakLossCount == 4 {
				multiple.Amount = math.Round(multiple.StartAmount*103300000*ratio/(3*0.99*(multiplier-1))) / 100000000
			}
		}

		if balanceChange == 0 {
			multiple.Status = 1
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
			return nil
		} else if balanceChange > 0 {
			multiple.WinAmount = math.Round((multiple.WinAmount+balanceChange)*10000000000) / 10000000000
			multiple.WinCount++
			multiple.StreakWinCount++
			multiple.StreakLossCount = 0
		} else {
			multiple.LossAmount = math.Round((multiple.LossAmount-balanceChange)*10000000000) / 10000000000
			multiple.LossCount++
			multiple.StreakWinCount = 0
			multiple.StreakLossCount++
		}

		betCount := multiple.WinCount + multiple.LossCount
		multiple.Rule = r.Switch(betCount)

		for _, interval := range intervals {
			if betCount%interval == 0 {
				multiple.Amount = math.Ceil((multiple.Amount+multiple.StartAmount)*100000000) / 100000000
			}
		}

		if multiple.BestProfit < multiple.Profit {
			multiple.BestProfit = multiple.Profit
		}

		if multiple.Profit < 0 && multiple.Invest < -multiple.Profit {
			multiple.Invest = -multiple.Profit
		}

		if multiple.Balance <= multiple.StopBalance+multiple.Amount {
			multiple.Status = 4
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
			r.Stop()
			return errors.New("stop loss")
		}

		if multiple.Balance >= multiple.TargetBalance {
			multiple.Status = 2
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
			r.Stop()
			return nil
		}

		multiple.Status = 1
		r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
	}

	for {
		if multiple.StreakLossCount >= 1 {
			time.Sleep(1 * time.Second)
		}

		if multiple.StreakLossCount >= 3 {
			time.Sleep(2 * time.Second)
		}

		if multiple.StreakLossCount >= 5 {
			time.Sleep(3 * time.Second)
		}

		if 2*multiple.WinCount < 3*multiple.LossCount {
			time.Sleep(2 * time.Second)
		}

		score, _ := r.Rdb.ZScore(
			r.Ctx,
			"wolf:strategies",
			"dice",
		).Result()
		if int64(score) != 3 {
			return errors.New("multiple not start")
		}

		hash, result, profit, err := r.Bet().Place(multiple.Currency, multiple.Rule, multiple.Amount, multiplier)
		if err != nil {
			multiple.Remark = err.Error()
			multiple.Status = 3
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
			return err
		}
		r.Hunt().Handing(hash, result)

		multiple.Balance = math.Round((multiple.Balance+profit)*10000000000) / 10000000000
		multiple.Profit = math.Round((multiple.Profit+profit)*10000000000) / 10000000000
		multiple.BufferProfit = math.Round((multiple.BufferProfit+profit)*10000000000) / 10000000000

		if multiple.BufferProfit >= 0 {
			multiple.Amount = multiple.StartAmount
			multiple.BufferProfit = 0
		} else {
			ratio := math.Round(multiple.BufferProfit / -multiple.StartAmount)
			if ratio > 0 && int(ratio)%7 == 0 || multiple.StreakLossCount == 4 {
				multiple.Amount = math.Round(multiple.StartAmount*103300000*ratio/(3*0.99*(multiplier-1))) / 100000000
			}
		}

		if profit >= 0 {
			multiple.WinAmount = math.Round((multiple.WinAmount+profit)*10000000000) / 10000000000
			multiple.WinCount++
			multiple.StreakWinCount++
			multiple.StreakLossCount = 0
		} else {
			multiple.LossAmount = math.Round((multiple.LossAmount-profit)*10000000000) / 10000000000
			multiple.LossCount++
			multiple.StreakWinCount = 0
			multiple.StreakLossCount++
		}

		betCount := multiple.WinCount + multiple.LossCount
		multiple.Rule = r.Switch(betCount)

		for _, interval := range intervals {
			if betCount%interval == 0 {
				multiple.Amount = math.Ceil((multiple.Amount+multiple.StartAmount)*100000000) / 100000000
			}
		}

		if multiple.BestProfit < multiple.Profit {
			multiple.BestProfit = multiple.Profit
		}

		if multiple.Profit < 0 && multiple.Invest < -multiple.Profit {
			multiple.Invest = -multiple.Profit
		}

		if multiple.Balance <= multiple.StopBalance+multiple.Amount {
			multiple.Status = 4
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
			r.Stop()
			return errors.New("stop loss")
		}

		if multiple.Balance >= multiple.TargetBalance {
			multiple.Status = 2
			r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
			r.Stop()
			break
		}

		r.Db.Model(&models.Multiple{ID: multiple.ID}).Select("*").Updates(multiple)
	}

	return nil
}

func (r *MultipleRepository) Rescue() error {
	var multiple models.Multiple
	result := r.Db.Where("status", 4).Take(&multiple)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("stop loss empty")
	}

	balance, err := r.Account().Balance(multiple.Currency)
	if err != nil {
		return err
	}

	if balance < multiple.StartBalance {
		return errors.New("balance not enough")
	}

	multiple.Balance = balance
	multiple.Status = 5
	r.Db.Model(&models.Multiple{ID: multiple.ID}).Updates(multiple)

	return nil
}

func (r *MultipleRepository) Clean() error {
	expire := time.Now().AddDate(1, 0, 0)
	r.Db.Where("updated_at < ?", expire).Delete(&models.Multiple{})
	return nil
}
