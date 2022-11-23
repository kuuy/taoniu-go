package dice

import (
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type PlansTask struct {
	Repository *repositories.PlansRepository
}

func (t *PlansTask) Apply(currency string) error {
	return t.Repository.Apply(currency)
}

func (t *PlansTask) Rescue() error {
	return t.Repository.Rescue()
}
