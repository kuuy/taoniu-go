package dice

import (
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type MultipleTask struct {
	Repository *repositories.MultipleRepository
}

func (t *MultipleTask) Apply(currency string) error {
	return t.Repository.Apply(currency)
}

func (t *MultipleTask) Rescue() error {
	return t.Repository.Rescue()
}
