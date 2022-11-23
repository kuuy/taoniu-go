package dice

import (
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type HellsTask struct {
	Repository *repositories.HellsRepository
}

func (t *HellsTask) Apply(currency string) error {
	return t.Repository.Apply(currency)
}

func (t *HellsTask) Rescue() error {
	return t.Repository.Rescue()
}
