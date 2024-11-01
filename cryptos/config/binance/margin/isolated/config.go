package isolated

const (
  REDIS_KEY_TRADINGS_LAST_PRICE       = "binance:margin.isolated:tradings:last:price:%v:%v"
  REDIS_KEY_TRADINGS_TRIGGERS_PLACE   = "binance:margin.isolated:tradings:triggers:place:%v:%v"
  SCALPING_MIN_BINANCE                = 50
  TRIGGERS_MIN_BINANCE                = 200
  LAUNCHPAD_MIN_BINANCE               = 1000
  LAUNCHPAD_DURATION                  = 30
  ASYNQ_QUEUE_TICKERS                 = "binance.margin.isolated.tickers"
  ASYNQ_QUEUE_KLINES                  = "binance.margin.isolated.klines"
  ASYNQ_QUEUE_DEPTH                   = "binance.margin.isolated.depth"
  ASYNQ_QUEUE_ACCOUNT                 = "binance.margin.isolated.account"
  ASYNQ_QUEUE_ORDERS                  = "binance.margin.isolated.orders"
  ASYNQ_QUEUE_INDICATORS              = "binance.margin.isolated.indicators"
  ASYNQ_QUEUE_STRATEGIES              = "binance.margin.isolated.strategies"
  ASYNQ_QUEUE_PLANS                   = "binance.margin.isolated.plans"
  ASYNQ_QUEUE_POSITIONS               = "binance.margin.isolated.positions"
  ASYNQ_QUEUE_TRADINGS_LAUNCHPAD      = "binance.margin.isolated.tradings.launchpad"
  ASYNQ_QUEUE_TRADINGS_SCALPING       = "binance.margin.isolated.tradings.scalping"
  ASYNQ_QUEUE_TRADINGS_TRIGGERS       = "binance.margin.isolated.tradings.triggers"
  ASYNQ_JOBS_ACCOUNT_FLUSH            = "binance:margin.isolated:account:flush"
  ASYNQ_JOBS_TICKERS_FLUSH            = "binance:margin.isolated:tickers:flush"
  ASYNQ_JOBS_TICKERS_UPDATE           = "binance:margin.isolated:tickers:update"
  ASYNQ_JOBS_KLINES_FLUSH             = "binance:margin.isolated:klines:flush"
  ASYNQ_JOBS_KLINES_UPDATE            = "binance:margin.isolated:klines:update"
  ASYNQ_JOBS_KLINES_CLEAN             = "binance:margin.isolated:klines:clean"
  ASYNQ_JOBS_TRADINGS_LAUNCHPAD_PLACE = "binance:margin.isolated:tradings:launchpad:place"
  ASYNQ_JOBS_TRADINGS_LAUNCHPAD_FLUSH = "binance:margin.isolated:tradings:launchpad:flush"
  ASYNQ_JOBS_TRADINGS_SCALPING_PLACE  = "binance:margin.isolated:tradings:scalping:place"
  ASYNQ_JOBS_TRADINGS_SCALPING_FLUSH  = "binance:margin.isolated:tradings:scalping:flush"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_PLACE  = "binance:margin.isolated:tradings:triggers:place"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_FLUSH  = "binance:margin.isolated:tradings:triggers:flush"
  NATS_INDICATORS_UPDATE              = "binance.margin.isolated.indicators.update"
  NATS_STRATEGIES_UPDATE              = "binance.margin.isolated.strategies.update"
  NATS_PLANS_UPDATE                   = "binance.margin.isolated.plans.update"
  NATS_ACCOUNT_UPDATE                 = "binance.margin.isolated.account.update"
  NATS_ORDERS_UPDATE                  = "binance.margin.isolated.orders.update"
  NATS_TICKERS_UPDATE                 = "binance.margin.isolated.tickers.update"
  NATS_KLINES_UPDATE                  = "binance.margin.isolated.klines.update"
  NATS_TRADINGS_SCALPING_PLACE        = "binance.margin.isolated.tradings.scalping.place"
  MQTT_TOPICS_ACCOUNT                 = "binance/margin.isolated/account/%s"
  MQTT_TOPICS_ORDERS                  = "binance/margin.isolated/orders/%s"
  MQTT_TOPICS_TICKERS                 = "binance/margin.isolated/tickers/%s"
  LOCKS_ACCOUNT_FLUSH                 = "locks:binance:margin.isolated:klines:flush"
  LOCKS_KLINES_FLUSH                  = "locks:binance:margin.isolated:klines:flush:%s:%s"
  LOCKS_KLINES_UPDATE                 = "locks:binance:margin.isolated:klines:update:%s:%s"
  LOCKS_KLINES_CLEAN                  = "locks:binance:margin.isolated:klines:clean:%s"
  LOCKS_KLINES_STREAM                 = "locks:binance:margin.isolated:klines:stream:%s:%s"
  LOCKS_ORDERS_OPEN                   = "locks:binance:margin.isolated:orders:open:%s"
  LOCKS_ORDERS_FLUSH                  = "locks:binance:margin.isolated:orders:flush:%s:%d"
  LOCKS_ORDERS_SYNC                   = "locks:binance:margin.isolated:orders:sync:%s"
  LOCKS_TRADINGS_PLACE                = "locks:binance:margin.isolated:tradings:place:%s"
  LOCKS_TRADINGS_LAUNCHPAD_PLACE      = "locks:binance:margin.isolated:tradings:launchpad:place:%s"
  LOCKS_TRADINGS_LAUNCHPAD_FLUSH      = "locks:binance:margin.isolated:tradings:launchpad:flush:%s"
  LOCKS_TRADINGS_SCALPING_PLACE       = "locks:binance:margin.isolated:tradings:scalping:place:%s"
  LOCKS_TRADINGS_SCALPING_FLUSH       = "locks:binance:margin.isolated:tradings:scalping:flush:%s"
  LOCKS_TRADINGS_TRIGGERS_PLACE       = "locks:binance:margin.isolated:tradings:triggers:place:%s"
  LOCKS_TRADINGS_TRIGGERS_FLUSH       = "locks:binance:margin.isolated:tradings:triggers:flush:%s"
)
