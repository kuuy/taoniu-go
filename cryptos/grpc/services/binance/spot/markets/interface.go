package markets

type TickersRepository interface {
  Gets(symbols []string, fields []string) []string
}
