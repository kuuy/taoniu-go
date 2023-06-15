package futures

type TradingsTriggersRepository interface {
  Scan() []string
}
