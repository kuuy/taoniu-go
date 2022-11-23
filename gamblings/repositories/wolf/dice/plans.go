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

type PlansRepository struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	UseProxy          bool
	BetRepository     *BetRepository
	HuntRepository    *HuntRepository
	AccountRepository *repositories.AccountRepository
}

func (r *PlansRepository) Bet() *BetRepository {
	if r.BetRepository == nil {
		r.BetRepository = &BetRepository{
			Rdb:      r.Rdb,
			Ctx:      r.Ctx,
			UseProxy: r.UseProxy,
		}
	}
	return r.BetRepository
}

func (r *PlansRepository) Hunt() *HuntRepository {
	if r.HuntRepository == nil {
		r.HuntRepository = &HuntRepository{
			Db:  r.Db,
			Rdb: r.Rdb,
			Ctx: r.Ctx,
		}
	}
	return r.HuntRepository
}

func (r *PlansRepository) Account() *repositories.AccountRepository {
	if r.AccountRepository == nil {
		r.AccountRepository = &repositories.AccountRepository{
			Rdb:      r.Rdb,
			Ctx:      r.Ctx,
			UseProxy: r.UseProxy,
		}
	}
	return r.AccountRepository
}

func (r *PlansRepository) Apply(currency string) error {
	score, _ := r.Rdb.ZScore(
		r.Ctx,
		"wolf:bet",
		"dice",
	).Result()
	if int64(score) == 0 {
		return errors.New("bet not started")
	}

	var plan models.Plan
	result := r.Db.Where("status IN ?", []int{1, 3, 4}).Take(&plan)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		result := r.Db.Order("updated_at desc").Take(&plan)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if time.Now().Unix()-plan.UpdatedAt.Unix() < 55 {
				return errors.New("plan next waiting")
			}
		}

		balance, err := r.Account().Balance(currency)
		if err != nil {
			return err
		}

		if balance < 1 {
			return errors.New("balance less than 1")
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

		amount := math.Round(balance*100000000*0.00000001) / 100000000

		rules := []string{"under", "over"}
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(rules), func(i, j int) { rules[i], rules[j] = rules[j], rules[i] })

		rule := rules[(rand.Intn(313-13)+13)%len(rules)]
		multiplier := 3.2174
		targetBalance := math.Round(balance*10000000000*1.00015) / 10000000000
		stopBalance := math.Round(balance*10000000000*0.998) / 10000000000

		if stopBalance < balance-10 {
			targetBalance = math.Round(balance*10000000000+20000000000) / 10000000000
			stopBalance = math.Round(balance*10000000000-100000000000) / 10000000000
		}

		plan = models.Plan{
			ID:              xid.New().String(),
			Currency:        currency,
			Rule:            rule,
			Amount:          amount,
			Multiplier:      multiplier,
			Balance:         balance,
			Invest:          amount,
			StartAmount:     amount,
			StartMultiplier: multiplier,
			StartBalance:    balance,
			TargetBalance:   targetBalance,
			StopBalance:     stopBalance,
			Status:          1,
		}
		r.Db.Create(&plan)
	} else {
		if plan.Status == 4 {
			return errors.New("stop loss occured")
		}
		if plan.Status == 3 {
			return errors.New("plan error waiting")
		}
		return errors.New("plan not finished")
	}

	r.Start()

	return nil
}

func (r *PlansRepository) Start() {
	r.Rdb.ZAdd(r.Ctx, "wolf:strategies", &redis.Z{
		Score:  2,
		Member: "dice",
	})
}

func (r *PlansRepository) Stop() {
	r.Rdb.ZRem(r.Ctx, "wolf:strategies", "dice")
}

func (r *PlansRepository) Switch(rule string, betCount int) string {
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

func (r *PlansRepository) Show() (models.Plan, error) {
	var plan models.Plan
	result := r.Db.Where("status IN ?", []int{1, 3}).Take(&plan)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return plan, errors.New("plan empty")
	}
	return plan, nil
}

func (r *PlansRepository) Place() error {
	var plan models.Plan
	result := r.Db.Where("status IN ?", []int{1, 3}).Take(&plan)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("plan empty")
	}

	if plan.Status == 3 {
		if time.Now().Unix()-plan.UpdatedAt.Unix() < 5 {
			return errors.New("plan error waiting")
		}

		balance, err := r.Account().Balance(plan.Currency)
		if err != nil {
			return err
		}

		balanceChange := balance - plan.Balance

		plan.Balance = balance
		plan.Profit = math.Round((balance-plan.StartBalance)*10000000000) / 10000000000

		if balanceChange == 0 {
			plan.Status = 1
			r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
			return nil
		} else if balanceChange > 0 {
			plan.Amount = plan.StartAmount
			plan.Multiplier = plan.StartMultiplier
			plan.WinAmount = math.Round((plan.WinAmount+balanceChange)*10000000000) / 10000000000
			plan.WinCount++
			plan.StreakWinCount++
			plan.StreakLossCount = 0
		} else {
			plan.Amount *= 2
			plan.Multiplier = math.Round(plan.Multiplier*0.975*10000) / 10000
			plan.LossAmount = math.Round((plan.LossAmount-balanceChange)*10000000000) / 10000000000
			plan.LossCount++
			plan.StreakWinCount = 0
			plan.StreakLossCount++
		}

		plan.Rule = r.Switch(plan.Rule, plan.WinCount+plan.LossCount)

		if plan.StreakLossCount >= 15 {
			plan.Amount = plan.StartAmount
			plan.Multiplier = plan.StartMultiplier
			plan.StreakLossCount = 0
		}

		if plan.BestProfit < plan.Profit {
			plan.BestProfit = plan.Profit
		}

		if plan.Profit < 0 {
			if plan.Invest < -plan.Profit {
				plan.Invest = -plan.Profit
			}
		}

		if plan.Balance <= plan.StopBalance+plan.Amount {
			plan.Status = 4
			r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
			r.Stop()
			return errors.New("stop loss")
		}

		if plan.Balance >= plan.TargetBalance {
			plan.Status = 2
			r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
			r.Stop()
			return nil
		}

		plan.Status = 1
		r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
	}

	for {
		if plan.StreakLossCount >= 6 {
			time.Sleep(1 * time.Second)
		}

		if plan.StreakLossCount >= 9 {
			time.Sleep(2 * time.Second)
		}

		if plan.StreakLossCount >= 12 {
			time.Sleep(3 * time.Second)
		}

		score, _ := r.Rdb.ZScore(
			r.Ctx,
			"wolf:strategies",
			"dice",
		).Result()
		if int64(score) != 2 {
			return errors.New("plan not start")
		}

		hash, result, profit, err := r.Bet().Place(plan.Currency, plan.Rule, plan.Amount, plan.Multiplier)
		if err != nil {
			plan.Remark = err.Error()
			plan.Status = 3
			r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
			return err
		}
		r.Hunt().Handing(hash, result)

		plan.Balance = math.Round((plan.Balance+profit)*10000000000) / 10000000000
		plan.Profit = math.Round((plan.Profit+profit)*10000000000) / 10000000000

		if profit >= 0 {
			plan.Amount = plan.StartAmount
			plan.Multiplier = plan.StartMultiplier
			plan.WinAmount = math.Round((plan.WinAmount+profit)*10000000000) / 10000000000
			plan.WinCount++
			plan.StreakWinCount++
			plan.StreakLossCount = 0
		} else {
			plan.Amount *= 2
			plan.Multiplier = math.Round(plan.Multiplier*0.975*10000) / 10000
			plan.LossAmount = math.Round((plan.LossAmount-profit)*10000000000) / 10000000000
			plan.LossCount++
			plan.StreakWinCount = 0
			plan.StreakLossCount++
		}

		plan.Rule = r.Switch(plan.Rule, plan.WinCount+plan.LossCount)

		if plan.StreakLossCount >= 15 {
			plan.Amount = plan.StartAmount
			plan.Multiplier = plan.StartMultiplier
			plan.StreakLossCount = 0
		}

		if plan.BestProfit < plan.Profit {
			plan.BestProfit = plan.Profit
		}

		if plan.Profit < 0 {
			if plan.Invest < -plan.Profit {
				plan.Invest = -plan.Profit
			}
		}

		if plan.Balance <= plan.StopBalance+plan.Amount {
			plan.Status = 4
			r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
			r.Stop()
			return errors.New("stop loss")
		}

		if plan.Balance >= plan.TargetBalance {
			plan.Status = 2
			r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
			break
		}

		r.Db.Model(&models.Plan{ID: plan.ID}).Select("*").Updates(plan)
	}

	return nil
}

func (r *PlansRepository) Rescue() error {
	var plan models.Plan
	result := r.Db.Where("status", 4).Take(&plan)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("stop loss empty")
	}

	balance, err := r.Account().Balance(plan.Currency)
	if err != nil {
		return err
	}

	if balance < plan.StartBalance {
		return errors.New("balance not enough")
	}

	plan.Balance = balance
	plan.Status = 5
	r.Db.Model(&models.Plan{ID: plan.ID}).Updates(plan)

	r.Start()

	return nil
}

func (r *PlansRepository) Clean() error {
	expire := time.Now().AddDate(1, 0, 0)
	r.Db.Where("updated_at < ?", expire).Delete(&models.Plan{})
	return nil
}
