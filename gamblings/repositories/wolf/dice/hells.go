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

type HellsRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	UseProxy          bool
	BetRepository     *BetRepository
	HuntRepository    *HuntRepository
	AccountRepository *repositories.AccountRepository
}

func (r *HellsRepository) Bet() *BetRepository {
	if r.BetRepository == nil {
		r.BetRepository = &BetRepository{
			Rdb:      r.Rdb,
			Ctx:      r.Ctx,
			UseProxy: r.UseProxy,
		}
	}
	return r.BetRepository
}

func (r *HellsRepository) Hunt() *HuntRepository {
	if r.HuntRepository == nil {
		r.HuntRepository = &HuntRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.HuntRepository
}

func (r *HellsRepository) Account() *repositories.AccountRepository {
	if r.AccountRepository == nil {
		r.AccountRepository = &repositories.AccountRepository{
			Rdb:      r.Rdb,
			Ctx:      r.Ctx,
			UseProxy: r.UseProxy,
		}
	}
	return r.AccountRepository
}

func (r *HellsRepository) Apply(currency string) error {
	score, _ := r.Rdb.ZScore(
		r.Ctx,
		"wolf:bet",
		"dice",
	).Result()
	if int64(score) == 0 {
		return errors.New("bet not started")
	}

	var hell models.Hell
	result := r.Db.Where("status IN ?", []int{1, 3, 4}).Take(&hell)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result := r.Db.Order("updated_at desc").Take(&hell)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if time.Now().Unix()-hell.UpdatedAt.Unix() < 55 {
				return errors.New("hell next waiting")
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
			return errors.New("bet place waiting")
		}

		balance, err := r.Account().Balance(currency)
		if err != nil {
			return err
		}

		if balance < 1 {
			return errors.New("balance less than 1")
		}

		amount := math.Round(balance*100000000*0.00000001) / 100000000

		rules := []string{"under", "over"}
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(rules), func(i, j int) { rules[i], rules[j] = rules[j], rules[i] })

		rule := rules[(rand.Intn(313-13)+13)%len(rules)]
		multiplier := 4.95
		targetBalance := math.Round(balance*10000000000*1.00028) / 10000000000
		stopBalance := math.Round(balance*10000000000*0.998) / 10000000000

		if stopBalance < balance-10 {
			targetBalance = math.Round(balance*10000000000+20000000000) / 10000000000
			stopBalance = math.Round(balance*10000000000-100000000000) / 10000000000
		}

		hell = models.Hell{
			ID:            xid.New().String(),
			Currency:      currency,
			Rule:          rule,
			Amount:        amount,
			Multiplier:    multiplier,
			Balance:       balance,
			Invest:        amount,
			StartAmount:   amount,
			StartBalance:  balance,
			TargetBalance: targetBalance,
			StopBalance:   stopBalance,
			Status:        1,
		}
		r.Db.Create(&hell)
	} else {
		if hell.Status == 4 {
			return errors.New("stop loss occured")
		}
		if hell.Status == 3 {
			return errors.New("hell error waiting")
		}
		return errors.New("hell not finished")
	}

	r.Start()

	return nil
}

func (r *HellsRepository) Start() {
	r.Rdb.ZAdd(r.Ctx, "wolf:strategies", &redis.Z{
		Score:  1,
		Member: "dice",
	})
}

func (r *HellsRepository) Stop() {
	r.Rdb.ZRem(r.Ctx, "wolf:strategies", "dice")
}

func (r *HellsRepository) Switch(rule string, betCount int) string {
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

func (r *HellsRepository) Show() (models.Hell, error) {
	var hell models.Hell
	result := r.Db.Where("status IN ?", []int{1, 3}).Take(&hell)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return hell, errors.New("hell empty")
	}
	return hell, nil
}

func (r *HellsRepository) Place() error {
	var hell models.Hell
	result := r.Db.Where("status IN ?", []int{1, 3}).Take(&hell)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("hell empty")
	}

	intervals := []int{1, 2, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40}

	if hell.Status == 3 {
		if time.Now().Unix()-hell.UpdatedAt.Unix() < 5 {
			return errors.New("hell error waiting")
		}

		balance, err := r.Account().Balance(hell.Currency)
		if err != nil {
			return err
		}

		balanceChange := balance - hell.Balance

		hell.Balance = balance
		hell.Profit = math.Round((balance-hell.StartBalance)*10000000000) / 10000000000

		if balanceChange == 0 {
			hell.Status = 1
			r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
			return nil
		} else if balanceChange > 0 {
			hell.Amount = hell.StartAmount
			hell.WinAmount = math.Round((hell.WinAmount+balanceChange)*10000000000) / 10000000000
			hell.WinCount++
			hell.StreakWinCount++
			hell.StreakLossCount = 0
		} else {
			hell.Amount = math.Ceil(hell.Amount*100000000*1.5) / 100000000
			hell.LossAmount = math.Round((hell.LossAmount-balanceChange)*10000000000) / 10000000000
			hell.LossCount++
			hell.StreakWinCount = 0
			hell.StreakLossCount++
		}

		betCount := hell.WinCount + hell.LossCount
		hell.Rule = r.Switch(hell.Rule, betCount)

		for _, interval := range intervals {
			if betCount%interval == 0 {
				hell.Amount = math.Ceil((hell.Amount+hell.StartAmount)*100000000) / 100000000
			}
		}

		if hell.BestProfit < hell.Profit {
			hell.BestProfit = hell.Profit
		}

		if hell.Profit < 0 {
			if hell.Invest < -hell.Profit {
				hell.Invest = -hell.Profit
			}
		}

		if hell.Balance <= hell.StopBalance+hell.Amount {
			hell.Status = 4
			r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
			r.Stop()
			return errors.New("stop loss")
		}

		if hell.Balance >= hell.TargetBalance {
			hell.Status = 2
			r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
			r.Stop()
			return nil
		}

		hell.Status = 1
		r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
	}

	for {
		if hell.StreakLossCount >= 10 {
			time.Sleep(1 * time.Second)
		}

		if hell.StreakLossCount >= 15 {
			time.Sleep(2 * time.Second)
		}

		if hell.StreakLossCount >= 20 {
			time.Sleep(3 * time.Second)
		}

		if hell.StreakLossCount >= 25 {
			time.Sleep(5 * time.Second)
		}

		score, _ := r.Rdb.ZScore(
			r.Ctx,
			"wolf:strategies",
			"dice",
		).Result()
		if int64(score) != 1 {
			return errors.New("hell not start")
		}

		hash, result, profit, err := r.Bet().Place(hell.Currency, hell.Rule, hell.Amount, hell.Multiplier)
		if err != nil {
			hell.Remark = err.Error()
			hell.Status = 3
			r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
			return err
		}
		r.Hunt().Handing(hash, result)

		hell.Balance = math.Round((hell.Balance+profit)*10000000000) / 10000000000
		hell.Profit = math.Round((hell.Profit+profit)*10000000000) / 10000000000

		if profit >= 0 {
			hell.Amount = hell.StartAmount
			hell.WinAmount = math.Round((hell.WinAmount+profit)*10000000000) / 10000000000
			hell.WinCount++
			hell.StreakWinCount++
			hell.StreakLossCount = 0
		} else {
			hell.Amount = math.Ceil(hell.Amount*100000000*1.5) / 100000000
			hell.LossAmount = math.Round((hell.LossAmount-profit)*10000000000) / 10000000000
			hell.LossCount++
			hell.StreakWinCount = 0
			hell.StreakLossCount++
		}

		betCount := hell.WinCount + hell.LossCount
		hell.Rule = r.Switch(hell.Rule, betCount)

		for _, interval := range intervals {
			if betCount%interval == 0 {
				hell.Amount = math.Ceil((hell.Amount+hell.StartAmount)*100000000) / 100000000
			}
		}

		if hell.BestProfit < hell.Profit {
			hell.BestProfit = hell.Profit
		}

		if hell.Profit < 0 {
			if hell.Invest < -hell.Profit {
				hell.Invest = -hell.Profit
			}
		}

		if hell.Balance <= hell.StopBalance+hell.Amount {
			hell.Status = 4
			r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
			r.Stop()
			return errors.New("stop loss")
		}

		if hell.Balance >= hell.TargetBalance {
			hell.Status = 2
			r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
			r.Stop()
			break
		}

		r.Db.Model(&models.Hell{ID: hell.ID}).Select("*").Updates(hell)
	}

	return nil
}

func (r *HellsRepository) Rescue() error {
	var hell models.Hell
	result := r.Db.Where("status", 4).Take(&hell)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("stop loss empty")
	}

	balance, err := r.Account().Balance(hell.Currency)
	if err != nil {
		return err
	}

	if balance < hell.StartBalance {
		return errors.New("balance not enough")
	}

	hell.Balance = balance
	hell.Status = 5
	r.Db.Model(&models.Hell{ID: hell.ID}).Updates(hell)

	r.Start()

	return nil
}

func (r *HellsRepository) Clean() error {
	expire := time.Now().AddDate(1, 0, 0)
	r.Db.Where("updated_at < ?", expire).Delete(&models.Hell{})
	return nil
}
