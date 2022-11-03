package models

import (
	"gorm.io/gorm"
	"taoniu.local/cryptos/models/account"
)

type Account struct{}

func NewAccount() *Account {
	return &Account{}
}

func (m *Account) AutoMigrate(db *gorm.DB) error {
	db.AutoMigrate(
		&account.User{},
	)

	return nil
}
