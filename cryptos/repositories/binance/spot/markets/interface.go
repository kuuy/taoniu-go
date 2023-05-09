package markets

type SymbolsRepository interface {
  Symbols() []string
}

type TickersRepository interface {
  Gets(symbols []string, fields []string) []string
}
