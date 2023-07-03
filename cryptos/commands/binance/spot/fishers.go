package spot

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/spot"
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
        Db: common.NewDB(),
      }
      h.Repository = &repositories.FishersRepository{
        Db: h.Db,
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
    },
  }
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

  //symbol := "ICPUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{8.893, 7.609, 6.798, 6.205, 5.791, 5.470, 5.158, 4.864, 4.569})
  //tickers = append(tickers, []float64{4.141, 3.614, 2.975})
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

  //symbol := "ADAUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.5543, 0.5001, 0.4606, 0.4309, 0.4071, 0.3879, 0.3772, 0.3569, 0.3444})
  //tickers = append(tickers, []float64{0.3311, 0.3175, 0.3027, 0.2819})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "LUNAUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{4.4712, 3.8313, 3.2680, 2.8646, 2.5307, 2.2809, 2.0679, 1.8957, 1.7148})
  //tickers = append(tickers, []float64{1.5718, 1.4580, 1.3354, 1.2157, 1.0790, 0.9237, 0.6799, 0.3957, 0.0342})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "LINKUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{10.700, 9.765, 9.017, 8.383, 7.896, 7.437, 6.978, 6.520, 6.095, 5.600})
  //tickers = append(tickers, []float64{5.025, 4.345, 3.411})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "XMRUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{180.4, 175.2, 171.0, 167.6, 164.8, 162.1, 159.6, 157.5, 155.4, 153.6})
  //tickers = append(tickers, []float64{151.5, 149.6, 147.2, 144.3, 140.5, 135.9, 130.2, 121.6})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "TRXUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.09869, 0.08937, 0.08334, 0.07872, 0.07488, 0.07214, 0.07013, 0.06830, 0.06660, 0.06501})
  //tickers = append(tickers, []float64{0.06337, 0.06172, 0.05984, 0.05734, 0.05417, 0.05040, 0.04498})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "KEYUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.012983, 0.009700, 0.007896, 0.006333, 0.004202})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "ARPAUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.10531, 0.08283, 0.06804, 0.05836, 0.04926, 0.03845, 0.02536})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "AMBUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.02410, 0.01803, 0.01392, 0.01187, 0.00773, 0.00356})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "LINAUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.036152, 0.022849, 0.016808, 0.013100, 0.009064, 0.004559, 0.001806})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "EPXUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.0004513, 0.0003752, 0.0003245, 0.0002956, 0.0002747, 0.0002556, 0.0002292})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "OGUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{14.346, 9.305, 6.441, 3.908, 2.107})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "BTCUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{45537.74, 33987.30, 28239.70, 25519.31, 23155.30, 20170.97, 15749.75})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "XRPUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{1.1750, 0.9454, 0.7995, 0.6970, 0.6120, 0.5412, 0.4942, 0.4594, 0.4316, 0.4083})
  //tickers = append(tickers, []float64{0.3871, 0.3653, 0.3346, 0.2903})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "DOTUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{17.439, 12.893, 10.733, 9.427, 8.784, 8.395, 7.687, 7.160, 6.824})
  //tickers = append(tickers, []float64{6.540, 6.288, 6.042, 5.778, 5.525, 5.255, 4.961, 4.633, 4.154})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "SHIBUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.00001669, 0.00001478, 0.00001338, 0.00001242, 0.00001165, 0.00001105, 0.00001045, 0.00000992, 0.00000927})
  //tickers = append(tickers, []float64{0.00000867, 0.00000792})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "APTUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{21.6108, 18.6161, 16.6705, 15.3087, 14.2256, 13.4568, 12.7212, 12.0211, 11.1682})
  //tickers = append(tickers, []float64{10.2188, 8.9154, 7.2817, 5.0749, 1.8796})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "DASHUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{81.91, 73.43, 67.98, 62.99, 58.55, 54.43, 50.27, 45.73, 39.98})
  //tickers = append(tickers, []float64{32.16, 20.64})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "XLMUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.1723, 0.1499, 0.1345, 0.1229, 0.1147, 0.1090, 0.1052, 0.1015, 0.0971})
  //tickers = append(tickers, []float64{0.0909, 0.0832, 0.0710})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "MASKSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{17.879, 14.135, 11.523, 9.483, 8.177, 7.142, 6.376, 5.822, 5.309})
  //tickers = append(tickers, []float64{4.510, 3.541, 2.359, 0.968})
  //return h.Repository.Apply(
  //	symbol,
  //	amount,
  //	balance,
  //	targetBalance,
  //	stopBalance,
  //	tickers,
  //)

  //symbol := "TRUUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{0.0490, 0.0431, 0.0386, 0.0345, 0.0311, 0.0266})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  //symbol := "EDUUSDT"
  //amount := 10.0
  //balance := 1500.0
  //targetBalance := 2900.0
  //stopBalance := 100.0
  //var tickers [][]float64
  //tickers = append(tickers, []float64{1.74623, 1.23508, 0.72157, 0.12914})
  //return h.Repository.Apply(
  //  symbol,
  //  amount,
  //  balance,
  //  targetBalance,
  //  stopBalance,
  //  tickers,
  //)

  symbol := "SUIUSDT"
  amount := 10.0
  balance := 1500.0
  targetBalance := 2900.0
  stopBalance := 100.0
  var tickers [][]float64
  tickers = append(tickers, []float64{1.3775, 0.9088, 0.5881, 0.0536})
  return h.Repository.Apply(
    symbol,
    amount,
    balance,
    targetBalance,
    stopBalance,
    tickers,
  )
}
