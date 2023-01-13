package repositories

import (
	"log"
)

type DhtRepository struct{}

func (r *DhtRepository) Crawl() error {
	log.Println("crawl")
	return nil
}
