package tradings

import (
	"log"
	repositories "taoniu.local/cryptos/repositories/binance/spot/margin/isolated/tradings"
)

type FIshersTask struct {
	Repository *repositories.FishersRepository
}

func (t *FIshersTask) Flush() error {
	symbols := t.Repository.Scan()
	for _, symbol := range symbols {
		err := t.Repository.Flush(symbol)
		if err != nil {
			log.Println("fishers flush error", err)
		}
	}
	return nil
}

func (t *FIshersTask) Place() error {
	symbols := t.Repository.Scan()
	for _, symbol := range symbols {
		err := t.Repository.Place(symbol)
		if err != nil {
			log.Println("fishers Place error", err)
		}
	}
	return nil
}
