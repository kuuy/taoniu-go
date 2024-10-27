package spot

const (
  REDIS_KEY_TRADINGS_LAST_PRICE                = "binance:spot:tradings:last:price:%v"
  REDIS_KEY_TRADINGS_TRIGGERS_PLACE            = "binance:spot:tradings:triggers:place:%v"
  SCALPING_MIN_BINANCE                         = 50
  TRIGGERS_MIN_BINANCE                         = 200
  LAUNCHPAD_MIN_BINANCE                        = 1000
  LAUNCHPAD_DURATION                           = 30
  ASYNQ_QUEUE_TICKERS                          = "binance.spot.tickers"
  ASYNQ_QUEUE_KLINES                           = "binance.spot.klines"
  ASYNQ_QUEUE_DEPTH                            = "binance.spot.depth"
  ASYNQ_QUEUE_ACCOUNT                          = "binance.spot.account"
  ASYNQ_QUEUE_ORDERS                           = "binance.spot.orders"
  ASYNQ_QUEUE_INDICATORS                       = "binance.spot.indicators"
  ASYNQ_QUEUE_STRATEGIES                       = "binance.spot.strategies"
  ASYNQ_QUEUE_PLANS                            = "binance.spot.plans"
  ASYNQ_QUEUE_POSITIONS                        = "binance.spot.positions"
  ASYNQ_QUEUE_TRADINGS_LAUNCHPAD               = "binance.spot.tradings.launchpad"
  ASYNQ_QUEUE_TRADINGS_SCALPING                = "binance.spot.tradings.scalping"
  ASYNQ_QUEUE_TRADINGS_TRIGGERS                = "binance.spot.tradings.triggers"
  ASYNQ_JOBS_ACCOUNT_FLUSH                     = "binance:spot:account:flush"
  ASYNQ_JOBS_TICKERS_FLUSH                     = "binance:spot:tickers:flush"
  ASYNQ_JOBS_TICKERS_UPDATE                    = "binance:spot:tickers:update"
  ASYNQ_JOBS_KLINES_FLUSH                      = "binance:spot:klines:flush"
  ASYNQ_JOBS_KLINES_UPDATE                     = "binance:spot:klines:update"
  ASYNQ_JOBS_KLINES_CLEAN                      = "binance:spot:klines:clean"
  ASYNQ_JOBS_ORDERS_OPEN                       = "binance:spot:orders:open"
  ASYNQ_JOBS_ORDERS_FLUSH                      = "binance:spot:orders:flush"
  ASYNQ_JOBS_ORDERS_SYNC                       = "binance:spot:orders:sync"
  ASYNQ_JOBS_TRADINGS_LAUNCHPAD_PLACE          = "binance:spot:tradings:launchpad:place"
  ASYNQ_JOBS_TRADINGS_LAUNCHPAD_FLUSH          = "binance:spot:tradings:launchpad:flush"
  ASYNQ_JOBS_TRADINGS_SCALPING_PLACE           = "binance:spot:tradings:scalping:place"
  ASYNQ_JOBS_TRADINGS_SCALPING_FLUSH           = "binance:spot:tradings:scalping:flush"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_PLACE           = "binance:spot:tradings:triggers:place"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_FLUSH           = "binance:spot:tradings:triggers:flush"
  NATS_INDICATORS_UPDATE                       = "binance.spot.indicators.update"
  NATS_STRATEGIES_UPDATE                       = "binance.spot.strategies.update"
  NATS_PLANS_UPDATE                            = "binance.spot.plans.update"
  NATS_ACCOUNT_UPDATE                          = "binance.spot.account.update"
  NATS_ORDERS_UPDATE                           = "binance.spot.orders.update"
  NATS_TICKERS_UPDATE                          = "binance.spot.tickers.update"
  NATS_KLINES_UPDATE                           = "binance.spot.klines.update"
  NATS_TRADINGS_SCALPING_PLACE                 = "binance.spot.tradings.scalping.place"
  MQTT_TOPICS_ACCOUNT                          = "binance/spot/account/%s"
  MQTT_TOPICS_ORDERS                           = "binance/spot/orders/%s"
  MQTT_TOPICS_TICKERS                          = "binance/spot/tickers/%s"
  LOCKS_ACCOUNT_FLUSH                          = "locks:binance:spot:account:flush"
  LOCKS_SYMBOLS_FLUSH                          = "locks:binance:spot:symbols:flush"
  LOCKS_KLINES_FLUSH                           = "locks:binance:spot:klines:flush:%s:%s"
  LOCKS_KLINES_UPDATE                          = "locks:binance:spot:klines:update:%s:%s"
  LOCKS_KLINES_CLEAN                           = "locks:binance:spot:klines:clean:%s"
  LOCKS_KLINES_STREAM                          = "locks:binance:spot:klines:stream:%s:%s"
  LOCKS_STRATEGIES_FLUSH                       = "locks:binance:spot:strategies:flush:%s:%s"
  LOCKS_ORDERS_OPEN                            = "locks:binance:spot:orders:open:%s"
  LOCKS_ORDERS_FLUSH                           = "locks:binance:spot:orders:flush:%s:%d"
  LOCKS_ORDERS_SYNC                            = "locks:binance:spot:orders:sync:%s"
  LOCKS_TRADINGS_PLACE                         = "locks:binance:spot:tradings:place:%s"
  LOCKS_TRADINGS_LAUNCHPAD_PLACE               = "locks:binance:spot:tradings:launchpad:place:%s"
  LOCKS_TRADINGS_LAUNCHPAD_FLUSH               = "locks:binance:spot:tradings:launchpad:flush:%s"
  LOCKS_TRADINGS_SCALPING_PLACE                = "locks:binance:spot:tradings:scalping:place:%s"
  LOCKS_TRADINGS_SCALPING_FLUSH                = "locks:binance:spot:tradings:scalping:flush:%s"
  LOCKS_TRADINGS_TRIGGERS_PLACE                = "locks:binance:spot:tradings:triggers:place:%s"
  LOCKS_TRADINGS_TRIGGERS_FLUSH                = "locks:binance:spot:tradings:triggers:flush:%s"
  LOCKS_TASKS_SYMBOLS_FLUSH                    = "locks:binance:spot:tasks:symbols:flush"
  LOCKS_TASKS_KLINES_FLUSH                     = "locks:binance:spot:tasks:klines:flush:%s"
  LOCKS_TASKS_KLINES_FIX                       = "locks:binance:spot:tasks:klines:fix:%s"
  LOCKS_TASKS_KLINES_CLEAN                     = "locks:binance:spot:tasks:klines:clean:%s"
  LOCKS_TASKS_STRATEGIES_CLEAN                 = "locks:binance:spot:tasks:strategies:clean:%s"
  LOCKS_TASKS_PLANS_CLEAN                      = "locks:binance:spot:tasks:plans:clean:%s"
  LOCKS_TASKS_ANALYSIS_TRADINGS_SCALPING_FLUSH = "locks:binance:spot:tasks:analysis:tradings:scalping:flush"
  LOCKS_TASKS_ANALYSIS_TRADINGS_TRIGGERS_FLUSH = "locks:binance:spot:tasks:analysis:tradings:triggers:flush"
)
