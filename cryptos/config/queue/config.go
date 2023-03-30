package queue

const (
	REDIS_ADDR                 = "127.0.0.1:6379"
	REDIS_DB                   = 11
	BINANCE_SPOT_DEPTH         = "cryptos.jobs.binance.spot.depth"
	BINANCE_SPOT_DEPTH_DELAY   = "cryptos.jobs.binance.spot.depth.delay"
	BINANCE_SPOT_TICKERS       = "cryptos.jobs.binance.spot.tickers"
	BINANCE_SPOT_TICKERS_DELAY = "cryptos.jobs.binance.spot.tickers.delay"
	BINANCE_SPOT_KLINES        = "cryptos.jobs.binance.spot.klines"
	BINANCE_SPOT_KLINES_DELAY  = "cryptos.jobs.binance.spot.klines.delay"
	TRADINGVIEW_ANALYSIS       = "tradingview.jobs.analysis"
	TRADINGVIEW_ANALYSIS_DELAY = "tradingview.jobs.analysis.delay"
)
