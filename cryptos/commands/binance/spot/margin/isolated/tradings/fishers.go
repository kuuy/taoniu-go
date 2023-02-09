package tradings

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
	tvRepositories "taoniu.local/cryptos/repositories/tradingview"
)

type FishersHandler struct {
	Db         *gorm.DB
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.FishersRepository
}

func NewFishersCommand() *cli.Command {
	var h FishersHandler
	return &cli.Command{
		Name:  "fishers",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = FishersHandler{
				Db:  common.NewDB(),
				Rdb: common.NewRedis(),
				Ctx: context.Background(),
			}
			h.Repository = &repositories.FishersRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.SymbolsRepository = &spotRepositories.SymbolsRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.AnalysisRepository = &tvRepositories.AnalysisRepository{
				Db: h.Db,
			}
			marginRepository := &spotRepositories.MarginRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.AccountRepository = marginRepository.Isolated().Account()
			h.Repository.OrdersRepository = marginRepository.Orders()
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:  "apply",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Apply(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
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
			{
				Name:  "place",
				Usage: "",
				Action: func(c *cli.Context) error {
					if err := h.Place(); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}
}

func (h *FishersHandler) Flush() error {
	symbols := h.Repository.Scan()
	for _, symbol := range symbols {
		err := h.Repository.Flush(symbol)
		if err != nil {
			log.Println("fishers flush error", err)
		}
	}
	return nil
}

func (h *FishersHandler) Place() error {
	symbols := h.Repository.Scan()
	for _, symbol := range symbols {
		err := h.Repository.Place(symbol)
		if err != nil {
			log.Println("fishers place error", err)
		}
	}
	return nil
}

func (h *FishersHandler) Apply() error {
	//symbol := "AVAXBUSD"
	//amount := 10.0
	//balance := 360.0
	//targetBalance := 400.0
	//stopBalance := 110.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{17.69, 17.38, 17.26, 16.95, 16.57, 16.43, 16.25, 16.02, 15.73}
	//tickers[1] = []float64{15.41, 15.23, 15.00, 14.88, 14.58, 14.38, 14.15, 14.03, 13.62}
	//tickers[2] = []float64{13.33, 13.1, 12.60, 12.52, 12.25, 12.06, 11.8, 11.61, 11.55}
	//tickers[3] = []float64{11.42, 11.35, 11.28, 11.14, 10.97, 10.89, 10.72, 10.61, 10.55}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "LAZIOBUSD"
	//amount := 10.0
	//balance := 360.0
	//targetBalance := 400.0
	//stopBalance := 110.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{3.5588, 3.5467, 3.4661, 3.3892, 3.3269, 3.272, 3.2243, 3.1731, 3.1401}
	//tickers[1] = []float64{3.0888, 2.9936, 2.9496, 2.9166, 2.88, 2.7774, 2.4477, 2.2756, 2.162}
	//tickers[2] = []float64{2.0741, 2.0265, 1.9678, 1.6345, 1.5576, 1.0557}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "DOGEBUSD"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{0.09215, 0.08958, 0.08758, 0.08579, 0.08369, 0.08084, 0.07999, 0.07742, 0.07515, 0.07192, 0.07049, 0.06765}
	//tickers[1] = []float64{0.06546, 0.06309, 0.05977, 0.05787, 0.05692, 0.05521, 0.05407, 0.05274, 0.05189, 0.05027}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "ALGOBUSD"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{0.2853, 0.2800, 0.2696, 0.2659, 0.2584, 0.2508, 0.241, 0.2355, 0.217}
	//tickers[1] = []float64{0.2088, 0.2025, 0.1976, 0.1923, 0.1886, 0.182, 0.1769, 0.1686, 0.1609}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "ZECBUSD"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{48.4, 47.4, 46.4, 45.7, 44.8, 44.3, 43.5, 42.5, 41.4}
	//tickers[1] = []float64{40.5, 39.7, 38.9, 38.2, 37.6, 36.7, 35.4}
	//return h.Repository.Apply(
	//  symbol,
	//  amount,
	//  balance,
	//  targetBalance,
	//  stopBalance,
	//  tickers,
	//)

	//symbol := "UNIBUSD"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{6.765, 6.649, 6.533, 6.407, 6.295, 6.198, 6.046, 5.921, 5.847}
	//tickers[1] = []float64{5.723, 5.628, 5.512, 5.429, 5.296, 5.210, 5.152, 5.082, 4.964}
	//tickers[2] = []float64{4.893, 4.764, 4.683}
	//return h.Repository.Apply(
	//  symbol,
	//  amount,
	//  balance,
	//  targetBalance,
	//  stopBalance,
	//  tickers,
	//)

	//symbol := "KAVABUSD"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{0.963, 0.941, 0.912, 0.894, 0.856, 0.834, 0.806, 0.763, 0.720}
	//tickers[1] = []float64{0.687, 0.655, 0.616, 0.576, 0.545, 0.524}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "MATICBUSD"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{1.2210, 1.1741, 1.1407, 1.0804, 1.0062, 0.9532, 0.9059, 0.8463, 0.7794}
	//tickers[1] = []float64{0.7360, 0.6992, 0.6573, 0.6186, 0.5867, 0.5489, 0.5000, 0.4681, 0.4319}
	//tickers[2] = []float64{0.4111, 0.3769, 0.3316}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "LTCBUSD"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{100.05, 98.50, 95.81, 94.17, 92.29, 90.42, 89.11, 86.90, 85.19}
	//tickers[1] = []float64{83.07, 80.37, 77.92, 75.47, 71.14, 68.86, 66.08, 62.32, 59.95}
	//tickers[2] = []float64{58.40, 55.54, 53.26, 50.56}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	symbol := "STPTBUSD"
	amount := 10.0
	balance := 500.0
	targetBalance := 900.0
	stopBalance := 100.0
	tickers := make([][]float64, 4)
	tickers[0] = []float64{0.04890, 0.04637, 0.04491, 0.04291, 0.04078, 0.03919, 0.03732, 0.03586, 0.03413}
	tickers[1] = []float64{0.03280, 0.03093, 0.02960, 0.02787, 0.02654, 0.02473}
	return h.Repository.Apply(
		symbol,
		amount,
		balance,
		targetBalance,
		stopBalance,
		tickers,
	)
}
