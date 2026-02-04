package swap

const (
  REDIS_KEY_INDICATORS = "raydium:swap:indicators:%v:%v:%v"
  REDIS_KEY_TICKERS    = "raydium:swap:realtime:%v"
  LOCKS_MINTS_FLUSH    = "locks:raydium:swap:mints:flush"
)
