package fishers

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
	repositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
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
			h.Repository.AccountRepository = &spotRepositories.AccountRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
			h.Repository.OrdersRepository = &spotRepositories.OrdersRepository{
				Db:  h.Db,
				Rdb: h.Rdb,
				Ctx: h.Ctx,
			}
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{41.22, 35.53, 29.58, 26.20, 23.33, 20.98, 19.12, 17.54, 16.30})
	//tickers = append(tickers, []float64{15.40, 14.56, 13.83, 13.10, 12.13, 11.12, 9.55})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{17.2781, 12.9592, 10.0692, 7.7945, 6.2292, 5.0049, 4.2636, 3.6513, 3.2275})
	//tickers = append(tickers, []float64{2.8906, 2.6432, 2.4600, 2.2821, 2.1204, 1.9504, 1.7825, 1.5473, 1.3049})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{0.14219, 0.12445, 0.11767, 0.11146, 0.10096, 0.09447, 0.08958, 0.08817, 0.08178})
	//tickers = append(tickers, []float64{0.07502, 0.06877, 0.06474, 0.06012, 0.05550, 0.05029})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{2.0438, 1.4962, 1.0609, 0.8820, 0.7661, 0.5744, 0.4341, 0.3655, 0.3155})
	//tickers = append(tickers, []float64{0.2879, 0.2442, 0.2218, 0.2082, 0.1962, 0.1804, 0.1619, 0.1373, 0.1060})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{113.5, 92.8, 77.6, 67.8, 59.3, 52.5, 48.9, 46.1, 43.5})
	//tickers = append(tickers, []float64{41.4, 39.4, 37.6, 35.6, 33.2, 30.6, 27.4, 23.9, 19.5})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{16.097, 13.871, 11.770, 9.985, 8.687, 7.833, 7.256, 6.832, 6.481})
	//tickers = append(tickers, []float64{6.189, 5.868, 5.560, 5.264, 4.860, 3.666, 2.944, 1.801})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{1.9245, 1.7354, 1.5916, 1.4789, 1.3738, 1.2937, 1.2300, 1.1741, 1.1407})
	//tickers = append(tickers, []float64{1.1099, 1.0425, 0.9747, 0.8847, 0.7722, 0.6190, 0.3892})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{273.64, 231.12, 197.25, 165.96, 144.94, 129.35, 116.12, 105.33, 94.27})
	//tickers = append(tickers, []float64{84.45, 75.66, 65.80, 49.16})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//tickers := make([][]float64, 2)
	//tickers[0] = []float64{0.07403, 0.05917, 0.05465, 0.05143, 0.04890, 0.04637, 0.04323, 0.03984, 0.03577}
	//tickers[1] = []float64{0.03014, 0.02471, 0.01395}
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{178.0, 156.2, 138.3, 123.7, 112.2, 101.2, 94.1, 86.8})
	//tickers = append(tickers, []float64{80.0, 74.6, 64.2, 53.0, 41.7, 26.3})
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
	//tickers = append(tickers, []float64{263.7, 213.3, 194.2, 176.7, 163.8, 158.8, 153.2, 144.2, 136.8})
	//tickers = append(tickers, []float64{132.0, 129.0, 126.1, 123.8, 121.5, 119.2, 116.8, 113.0, 111.3})
	//tickers = append(tickers, []float64{109.3, 102.0, 94.4, 83.4})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{16.405, 15.374, 14.857, 14.436, 14.114, 13.811, 13.540, 13.310, 13.066})
	//tickers = append(tickers, []float64{12.875, 12.689, 12.496, 12.221, 11.948, 11.548, 11.189, 10.953, 10.239})
	//tickers = append(tickers, []float64{9.345, 8.292})
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
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{0.17395, 0.13905, 0.11244, 0.09151, 0.07463, 0.06178, 0.05162, 0.04465, 0.03976})
	//tickers = append(tickers, []float64{0.03656, 0.03315, 0.02931, 0.02931, 0.02428, 0.01962, 0.01473})
	//
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

	//symbol := "QTUMUSDT"
	//amount := 10.0
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{7.297, 6.442, 5.727, 5.058, 4.491, 4.057, 3.641, 3.206, 2.979})
	//tickers = append(tickers, []float64{2.852, 2.745, 2.668, 2.585, 2.519, 2.449, 2.198, 1.899, 1.310})
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)

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

	symbol := "ICPUSDT"
	amount := 10.0
	balance := 1500.0
	targetBalance := 2900.0
	stopBalance := 100.0
	var tickers [][]float64
	tickers = append(tickers, []float64{8.893, 7.609, 6.798, 6.205, 5.791, 5.470, 5.158, 4.864, 4.569})
	tickers = append(tickers, []float64{4.141, 3.614, 2.975})
	return h.Repository.Apply(
		symbol,
		amount,
		balance,
		targetBalance,
		stopBalance,
		tickers,
	)

	//symbol := "WOOUSDT"
	//amount := 10.0
	//balance := 1500.0
	//targetBalance := 2900.0
	//stopBalance := 100.0
	//var tickers [][]float64
	//tickers = append(tickers, []float64{0.3366, 0.2989, 0.2728, 0.2515, 0.2346, 0.2202, 0.2081, 0.1962, 0.1842})
	//tickers = append(tickers, []float64{0.1691, 0.1497, 0.1224, 0.0879})
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
	//tickers = append(tickers, []float64{63.55, 55.44, 45.61, 39.84, 34.94, 31.05, 27.92, 25.10, 23.19})
	//tickers = append(tickers, []float64{21.22, 19.19, 18.27, 14.13, 8.75})
	//return h.Repository.Apply(
	//	symbol,
	//	amount,
	//	balance,
	//	targetBalance,
	//	stopBalance,
	//	tickers,
	//)
}
