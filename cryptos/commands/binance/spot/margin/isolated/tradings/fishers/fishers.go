package fishers

import (
	"context"
	"log"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings/fishers"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
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
	//symbol := "AVAXUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 5)
	//tickers[0] = []float64{21.95, 20.65, 20.08, 19.58, 19.28, 18.84, 17.69, 17.38, 17.26}
	//tickers[1] = []float64{16.95, 16.57, 16.43, 16.25, 16.15, 16.06, 15.96, 15.88, 15.73}
	//tickers[2] = []float64{15.41, 15.23, 15.00, 14.88, 14.58, 14.38, 14.15, 14.03, 13.62}
	//tickers[3] = []float64{13.33, 13.1, 12.60, 12.52, 12.25, 12.06, 11.8, 11.61, 11.55}
	//tickers[4] = []float64{11.42, 11.35, 11.28, 11.14, 10.97, 10.89, 10.72, 10.61, 10.55}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "LAZIOUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{4.2617, 4.1064, 3.9549, 3.7801, 3.6636, 3.5588, 3.5467, 3.4661, 3.3892}
	//tickers[1] = []float64{3.3269, 3.272, 3.2243, 3.1731, 3.1401, 3.0888, 2.9936, 2.9496}
	//tickers[2] = []float64{2.9166, 2.88, 2.7774, 2.4477, 2.2756, 2.162, 2.0741, 2.0265, 1.9678}
	//tickers[3] = []float64{1.6345, 1.5576, 1.0557}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "DOGEUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 3)
	//tickers[0] = []float64{0.11767, 0.11264, 0.10808, 0.10328, 0.09567, 0.09215, 0.08958, 0.08758, 0.08579, 0.08369}
	//tickers[1] = []float64{0.08084, 0.07999, 0.07742, 0.07515, 0.07286, 0.07049, 0.06765, 0.06546, 0.06309, 0.05977}
	//tickers[2] = []float64{0.05787, 0.05692, 0.05521, 0.05407, 0.05274, 0.05189, 0.05027}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "ALGOUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{0.3243, 0.3148, 0.3054, 0.2951, 0.2853, 0.2800, 0.2734, 0.2659, 0.2584})
	//tickers = append(tickers, []float64{0.2508, 0.241, 0.2355, 0.217, 0.2088, 0.2025, 0.1976, 0.1923, 0.1886})
	//tickers = append(tickers, []float64{0.182, 0.1769, 0.1686, 0.1609})
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "ZECUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 3)
	//tickers[0] = []float64{53.0, 51.6, 50.1, 48.4, 47.4, 46.4, 45.7, 44.8, 44.3}
	//tickers[1] = []float64{43.5, 42.5, 41.4, 40.5, 39.7, 38.9, 38.2, 37.6, 36.7}
	//tickers[2] = []float64{35.4}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "UNIUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 3)
	//tickers[0] = []float64{7.949, 7.618, 7.359, 7.118, 6.927, 6.765, 6.649, 6.533, 6.407}
	//tickers[1] = []float64{6.295, 6.198, 6.046, 5.921, 5.847, 5.723, 5.628, 5.512, 5.429}
	//tickers[2] = []float64{5.296, 5.210, 5.152, 5.082, 4.964, 4.893, 4.764, 4.683}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "KAVAUSDT"
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

	//symbol := "MATICUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 4)
	//tickers[0] = []float64{1.6723, 1.6064, 1.5363, 1.4580, 1.3797, 1.3159, 1.2778, 1.2479, 1.2210}
	//tickers[1] = []float64{1.1741, 1.1407, 1.0804, 1.0062, 0.9532, 0.9059, 0.8463, 0.7794, 0.7360}
	//tickers[2] = []float64{0.6992, 0.6573, 0.6186, 0.5867, 0.5489, 0.5000, 0.4681, 0.4319, 0.4111}
	//tickers[3] = []float64{0.3769, 0.3316}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "LTCUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 3)
	//tickers[0] = []float64{111.35, 109.32, 106.99, 104.09, 101.05, 98.50, 95.81, 94.17, 92.29}
	//tickers[1] = []float64{90.42, 89.11, 86.90, 85.19, 83.07, 80.37, 77.92, 75.47, 71.14}
	//tickers[2] = []float64{68.86, 66.08, 62.32, 59.95, 58.40, 55.54, 53.26, 50.56}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "STPTUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 2)
	//tickers[0] = []float64{0.05748, 0.05459, 0.05143, 0.04890, 0.04637, 0.04491, 0.04291, 0.04078, 0.03919}
	//tickers[1] = []float64{0.03732, 0.03586, 0.03413, 0.03280, 0.03093, 0.02960, 0.02787, 0.02654, 0.02473}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "AAVEUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 3)
	//tickers[0] = []float64{104.8, 100.5, 96.0, 92.2, 88.5, 85.3, 83.2, 81.3, 79.8}
	//tickers[1] = []float64{77.8, 76.1, 74.5, 72.2, 70.0, 67.9, 63.9, 62.2, 60.4}
	//tickers[2] = []float64{58.6, 56.3, 54.3, 52.4, 50.6, 48.6, 46.9}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "BCHUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{151.6, 145.9, 140.0, 136.5, 132.1, 130.4, 129.0, 127.2, 125.6})
	//tickers = append(tickers, []float64{123.8, 122.0, 120.5, 118.8, 116.8, 115.1, 113.0, 111.3, 109.3})
	//tickers = append(tickers, []float64{107.8, 106.0, 104.3, 102.0, 100.0, 98.4, 96.4, 94.4, 92.6})
	//tickers = append(tickers, []float64{91.0, 89.0})
	//
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "CFXUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 5)
	//tickers[0] = []float64{0.1430, 0.1357, 0.1272, 0.1196, 0.1114, 0.1047, 0.0965, 0.8889, 0.0811}
	//tickers[1] = []float64{0.0765, 0.0729, 0.0685, 0.0658, 0.0639, 0.0620, 0.0604, 0.0582, 0.0563}
	//tickers[2] = []float64{0.0547, 0.0530, 0.0510, 0.0508, 0.0493, 0.0476, 0.0454, 0.0434, 0.0411}
	//tickers[3] = []float64{0.0389, 0.0368, 0.0346, 0.0325, 0.0310, 0.0296, 0.0276, 0.0265, 0.0257}
	//tickers[4] = []float64{0.0249, 0.0236, 0.0223}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "ATOMUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 5)
	//tickers[0] = []float64{15.374, 14.857, 14.530, 14.114, 13.870, 13.540, 13.310, 13.098, 13.066}
	//tickers[1] = []float64{12.870, 12.718, 12.446, 12.213, 11.985, 11.772, 11.536, 11.244, 10.953}
	//tickers[2] = []float64{10.626, 10.232, 9.818, 9.343, 8.585}
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "ANKRUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{0.05919, 0.05045, 0.04476, 0.04049, 0.03728, 0.03602, 0.03470, 0.03344, 0.03212})
	//tickers = append(tickers, []float64{0.03087, 0.02997, 0.02911, 0.02823, 0.02762, 0.02715, 0.02705, 0.02636, 0.02571})
	//tickers = append(tickers, []float64{0.02499, 0.02426, 0.02343, 0.02266, 0.02125, 0.02064, 0.02011, 0.01960})
	//
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	symbol := "QTUMUSDT"
	amount := 10.0
	balance := 500.0
	targetBalance := 900.0
	stopBalance := 100.0
	var tickers [][]float64
	tickers = append(tickers, []float64{5.489, 4.809, 4.262, 3.631, 3.302, 3.153, 3.031, 2.894, 2.815})
	tickers = append(tickers, []float64{2.741, 2.633, 2.542, 2.450, 2.368, 2.290, 2.225, 2.151, 2.055})
	tickers = append(tickers, []float64{1.964, 1.873, 1.782})
	return h.Repository.Apply(
		symbol,
		amount,
		balance,
		targetBalance,
		stopBalance,
		tickers,
	)

	//symbol := "BNBUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{374.9, 343.8, 319.4, 313.8, 309.0, 304.4, 300.3, 296.2, 292.5})
	//tickers = append(tickers, []float64{289.3, 287.0, 283.6, 280.1, 276.6, 273.0, 268.7})
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "ICPUSDT"
	//amount := 10.0
	//balance := 500.0
	//targetBalance := 900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{1000.0})
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "WOOUSDT"
	//amount := 10.0
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{0.2602, 0.2517, 0.2420, 0.2312, 0.2193, 0.2081, 0.1962, 0.1853, 0.1735})
	//tickers = append(tickers, []float64{0.1620, 0.1489, 0.1379, 0.1291, 0.1212, 0.1153, 0.1095, 0.1034})
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "SOLUSDT"
	//amount := 10.0
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{36.00, 30.91, 28.23, 26.37, 25.06, 23.54, 21.94, 20.40, 19.28})
	//tickers = append(tickers, []float64{18.27, 17.28, 16.35, 15.44, 14.55, 13.73, 12.97, 12.53, 11.90})
	//tickers = append(tickers, []float64{11.41, 11.01, 9.77, 8.75})
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)
}
