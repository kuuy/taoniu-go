package spot

const (
  REDIS_KEY_BALANCE                            = "binance:spot:balance:%v"
  REDIS_KEY_TRADINGS_LAST_PRICE                = "binance:spot:tradings:last:price:%v"
  REDIS_KEY_TRADINGS_TRIGGERS_PLACE            = "binance:spot:tradings:triggers:place:%v"
  REDIS_KEY_TRADINGS_GAMBLING_SCALPING_PLACE   = "binance:spot:tradings:gambling:scalping:place:%v"
  SCALPING_MIN_BINANCE                         = 50
  TRIGGERS_MIN_BINANCE                         = 200
  LAUNCHPAD_MIN_BINANCE                        = 1000
  LAUNCHPAD_DURATION                           = 30
  GAMBLING_SCALPING_MIN_BINANCE                = 2121
  GAMBLING_SCALPING_MIN_AMOUNT                 = 16
  GAMBLING_SCALPING_MAX_AMOUNT                 = 440
  GAMBLING_SCALPING_PRICE_LOSE_PERCENT         = 20
  GAMBLING_ANT_MIN_BINANCE                     = 3000
  GAMBLING_ANT_MAX_AMOUNT                      = 125
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
  ASYNQ_QUEUE_TRADINGS_GAMBLING                = "binance.spot.tradings.gambling"
  ASYNQ_JOBS_ACCOUNT_FLUSH                     = "binance:spot:account:flush"
  ASYNQ_JOBS_TICKERS_FLUSH                     = "binance:spot:tickers:flush"
  ASYNQ_JOBS_TICKERS_UPDATE                    = "binance:spot:tickers:update"
  ASYNQ_JOBS_KLINES_FLUSH                      = "binance:spot:klines:flush"
  ASYNQ_JOBS_KLINES_UPDATE                     = "binance:spot:klines:update"
  ASYNQ_JOBS_KLINES_CLEAN                      = "binance:spot:klines:clean"
  ASYNQ_JOBS_INDICATORS_ATR                    = "binance:spot:indicators:atr"
  ASYNQ_JOBS_INDICATORS_ZLEMA                  = "binance:spot:indicators:zlema"
  ASYNQ_JOBS_INDICATORS_HA_ZLEMA               = "binance:spot:indicators:ha_zlema"
  ASYNQ_JOBS_INDICATORS_KDJ                    = "binance:spot:indicators:kdj"
  ASYNQ_JOBS_INDICATORS_BBANDS                 = "binance:spot:indicators:bbands"
  ASYNQ_JOBS_INDICATORS_ICHIMOKU_CLOUD         = "binance:spot:indicators:ichimoku_cloud"
  ASYNQ_JOBS_INDICATORS_PIVOT                  = "binance:spot:indicators:pivot"
  ASYNQ_JOBS_INDICATORS_VOLUME_PROFILE         = "binance:spot:indicators:volume_profile"
  ASYNQ_JOBS_INDICATORS_ANDEAN_OSCILLATOR      = "binance:spot:indicators:andean_oscillator"
  ASYNQ_JOBS_STRATEGIES_ATR                    = "binance:spot:strategies:atr"
  ASYNQ_JOBS_STRATEGIES_ZLEMA                  = "binance:spot:strategies:zlema"
  ASYNQ_JOBS_STRATEGIES_HA_ZLEMA               = "binance:spot:strategies:ha_zlema"
  ASYNQ_JOBS_STRATEGIES_KDJ                    = "binance:spot:strategies:kdj"
  ASYNQ_JOBS_STRATEGIES_BBANDS                 = "binance:spot:strategies:bbands"
  ASYNQ_JOBS_STRATEGIES_ICHIMOKU_CLOUD         = "binance:spot:strategies:ichimoku_cloud"
  ASYNQ_JOBS_ORDERS_OPEN                       = "binance:spot:orders:open"
  ASYNQ_JOBS_ORDERS_FLUSH                      = "binance:spot:orders:flush"
  ASYNQ_JOBS_ORDERS_SYNC                       = "binance:spot:orders:sync"
  ASYNQ_JOBS_TRADINGS_LAUNCHPAD_PLACE          = "binance:spot:tradings:launchpad:place"
  ASYNQ_JOBS_TRADINGS_LAUNCHPAD_FLUSH          = "binance:spot:tradings:launchpad:flush"
  ASYNQ_JOBS_TRADINGS_SCALPING_PLACE           = "binance:spot:tradings:scalping:place"
  ASYNQ_JOBS_TRADINGS_SCALPING_FLUSH           = "binance:spot:tradings:scalping:flush"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_PLACE           = "binance:spot:tradings:triggers:place"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_FLUSH           = "binance:spot:tradings:triggers:flush"
  ASYNQ_JOBS_TRADINGS_GAMBLING_ANT_PLACE       = "binance:spot:tradings:gambling:ant:place"
  ASYNQ_JOBS_TRADINGS_GAMBLING_ANT_FLUSH       = "binance:spot:tradings:gambling:ant:flush"
  ASYNQ_JOBS_TRADINGS_GAMBLING_SCALPING_PLACE  = "binance:spot:tradings:gambling:scalping:place"
  NATS_INDICATORS_UPDATE                       = "binance.spot.indicators.update"
  NATS_STRATEGIES_UPDATE                       = "binance.spot.strategies.update"
  NATS_PLANS_UPDATE                            = "binance.spot.plans.update"
  NATS_ACCOUNT_UPDATE                          = "binance.spot.account.update"
  NATS_ORDERS_UPDATE                           = "binance.spot.orders.update"
  NATS_TICKERS_UPDATE                          = "binance.spot.tickers.update"
  NATS_KLINES_UPDATE                           = "binance.spot.klines.update"
  NATS_TRADINGS_SCALPING_PLACE                 = "binance.spot.tradings.scalping.place"
  MQTT_TOPICS_ACCOUNT                          = "binance/spot/account/%v"
  MQTT_TOPICS_ORDERS                           = "binance/spot/orders/%v"
  MQTT_TOPICS_TICKERS                          = "binance/spot/tickers/%v"
  LOCKS_ACCOUNT_FLUSH                          = "locks:binance:spot:account:flush"
  LOCKS_SYMBOLS_FLUSH                          = "locks:binance:spot:symbols:flush"
  LOCKS_KLINES_FLUSH                           = "locks:binance:spot:klines:flush:%v:%v"
  LOCKS_KLINES_UPDATE                          = "locks:binance:spot:klines:update:%v:%v"
  LOCKS_KLINES_CLEAN                           = "locks:binance:spot:klines:clean:%v"
  LOCKS_KLINES_STREAM                          = "locks:binance:spot:klines:stream:%v:%v"
  LOCKS_INDICATORS_FLUSH                       = "locks:binance:spot:indicators:flush:%v:%v"
  LOCKS_STRATEGIES_FLUSH                       = "locks:binance:spot:strategies:flush:%v:%v"
  LOCKS_ORDERS_OPEN                            = "locks:binance:spot:orders:open:%v"
  LOCKS_ORDERS_FLUSH                           = "locks:binance:spot:orders:flush:%v:%d"
  LOCKS_ORDERS_SYNC                            = "locks:binance:spot:orders:sync:%v"
  LOCKS_TRADINGS_PLACE                         = "locks:binance:spot:tradings:place:%v"
  LOCKS_TRADINGS_TAKE                          = "locks:binance:spot:tradings:take:%v"
  LOCKS_TRADINGS_LAUNCHPAD_PLACE               = "locks:binance:spot:tradings:launchpad:place:%v"
  LOCKS_TRADINGS_LAUNCHPAD_FLUSH               = "locks:binance:spot:tradings:launchpad:flush:%v"
  LOCKS_TRADINGS_SCALPING_PLACE                = "locks:binance:spot:tradings:scalping:place:%v"
  LOCKS_TRADINGS_SCALPING_FLUSH                = "locks:binance:spot:tradings:scalping:flush:%v"
  LOCKS_TRADINGS_TRIGGERS_PLACE                = "locks:binance:spot:tradings:triggers:place:%v"
  LOCKS_TRADINGS_TRIGGERS_FLUSH                = "locks:binance:spot:tradings:triggers:flush:%v"
  LOCKS_TRADINGS_GAMBLING_ANT_PLACE            = "locks:binance:spot:tradings:gambling:ant:place:%v"
  LOCKS_TRADINGS_GAMBLING_ANT_FLUSH            = "locks:binance:spot:tradings:gambling:ant:flush:%v"
  LOCKS_TASKS_SYMBOLS_FLUSH                    = "locks:binance:spot:tasks:symbols:flush"
  LOCKS_TASKS_KLINES_FLUSH                     = "locks:binance:spot:tasks:klines:flush:%v"
  LOCKS_TASKS_KLINES_FIX                       = "locks:binance:spot:tasks:klines:fix:%v"
  LOCKS_TASKS_KLINES_CLEAN                     = "locks:binance:spot:tasks:klines:clean:%v"
  LOCKS_TASKS_STRATEGIES_CLEAN                 = "locks:binance:spot:tasks:strategies:clean:%v"
  LOCKS_TASKS_PLANS_CLEAN                      = "locks:binance:spot:tasks:plans:clean:%v"
  LOCKS_TASKS_ANALYSIS_TRADINGS_SCALPING_FLUSH = "locks:binance:spot:tasks:analysis:tradings:scalping:flush"
  LOCKS_TASKS_ANALYSIS_TRADINGS_TRIGGERS_FLUSH = "locks:binance:spot:tasks:analysis:tradings:triggers:flush"
)
