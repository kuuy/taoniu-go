package spot

import (
	"context"
	"github.com/hibiken/asynq"
	"log"
	config "taoniu.local/cryptos/config/queue"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	tasks "taoniu.local/cryptos/queue/tasks/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot"
)

type TickersHandler struct {
	Db                *gorm.DB
	Rdb               *redis.Client
	Ctx               context.Context
	Task              *tasks.TickersTask
	Repository        *repositories.TickersRepository
	SymbolsRepository *repositories.SymbolsRepository
}

func NewTickersCommand() *cli.Command {
	var h TickersHandler
	return &cli.Command{
		Name:  "tickers",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = TickersHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Task = &tasks.TickersTask{}
			h.Repository = &repositories.TickersRepository{
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.SymbolsRepository = &repositories.SymbolsRepository{
				Db: h.Db,
			}
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "flush",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Flush(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *TickersHandler) Flush() error {
	log.Println("Tickers flush...")
	//symbols := h.SymbolsRepository.Scan()
	//log.Println(symbols)
	//for i := 0; i < len(symbols); i += 20 {
	//	j := i + 20
	//	if j > len(symbols) {
	//		j = len(symbols)
	//	}
	//	h.Repository.Flush(symbols[i:j])
	//}
	rdb := asynq.RedisClientOpt{
		Addr: config.REDIS_ADDR,
		DB:   config.REDIS_DB,
	}
	client := asynq.NewClient(rdb)
	defer client.Close()
	task, err := h.Task.Flush([]string{"ADAUSDT", "AVAXUSDT"})
	if err != nil {
		return err
	}
	info, err := client.Enqueue(
		task,
		asynq.Queue(config.BINANCE_SPOT_TICKERS),
		asynq.MaxRetry(0),
		asynq.Timeout(3*time.Minute),
	)
	if err != nil {
		return nil
	}
	log.Println("task", task.Type(), info.ID, info.Queue)

	return nil
}
