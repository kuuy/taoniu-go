package wolf

import (
	repositories "taoniu.local/gamblings/repositories/wolf"
)

type AccountTask struct {
	Repository *repositories.AccountRepository
}

func (t *AccountTask) Flush() error {
	t.Repository.Balance("")
	return nil
}
