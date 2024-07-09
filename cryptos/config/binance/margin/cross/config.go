package cross

const (
  SCALPING_MAX_BORROWED              = 10000
  TRIGGERS_MAX_BORROWED              = 10000
  ASYNQ_QUEUE_ACCOUNT                = "binance.margin.cross.account"
  ASYNQ_QUEUE_ORDERS                 = "binance.margin.cross.orders"
  ASYNQ_QUEUE_POSITIONS              = "binance.margin.cross.positions"
  ASYNQ_QUEUE_TRADINGS_SCALPING      = "binance.margin.cross.tradings.scalping"
  ASYNQ_QUEUE_TRADINGS_TRIGGERS      = "binance.margin.cross.tradings.triggers"
  ASYNQ_JOBS_ACCOUNT_FLUSH           = "binance:margin:cross:account:flush"
  ASYNQ_JOBS_TRADINGS_SCALPING_PLACE = "binance:margin:cross:tradings:scalping:place"
  ASYNQ_JOBS_TRADINGS_SCALPING_FLUSH = "binance:margin:cross:tradings:scalping:flush"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_PLACE = "binance:margin:cross:tradings:triggers:place"
  ASYNQ_JOBS_TRADINGS_TRIGGERS_FLUSH = "binance:margin:cross:tradings:triggers:flush"
  NATS_ACCOUNT_UPDATE                = "binance.margin.cross.account.update"
  NATS_ORDERS_UPDATE                 = "binance.margin.cross.orders.update"
  NATS_TICKERS_UPDATE                = "binance.margin.cross.tickers.update"
  NATS_TRADINGS_SCALPING_PLACE       = "binance.margin.cross.tradings.scalping.place"
  MQTT_TOPICS_ACCOUNT                = "binance/margin/cross/account/%s"
  MQTT_TOPICS_ORDERS                 = "binance/margin/cross/orders/%s"
  LOCKS_ACCOUNT_FLUSH                = "locks:binance:margin:cross:account:flush"
  LOCKS_ACCOUNT_BORROW               = "locks:binance:margin:cross:account:borrow:%s"
  LOCKS_ORDERS_OPEN                  = "locks:binance:margin:cross:orders:open:%s"
  LOCKS_ORDERS_FLUSH                 = "locks:binance:margin:cross:orders:flush:%s:%d"
  LOCKS_ORDERS_SYNC                  = "locks:binance:margin:cross:orders:sync:%s"
  LOCKS_TRADINGS_PLACE               = "locks:binance:margin:cross:tradings:place:%s"
  LOCKS_TRADINGS_SCALPING_PLACE      = "locks:binance:margin:cross:tradings:scalping:place:%s"
  LOCKS_TRADINGS_SCALPING_FLUSH      = "locks:binance:margin:cross:tradings:scalping:flush:%s"
  LOCKS_TRADINGS_TRIGGERS_PLACE      = "locks:binance:margin:cross:tradings:triggers:place:%s"
  LOCKS_TRADINGS_TRIGGERS_FLUSH      = "locks:binance:margin:cross:tradings:triggers:flush:%s"
)
