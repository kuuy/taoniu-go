package dice

import (
	repositories "taoniu.local/gamblings/repositories/wolf/dice"
)

type MultipleTask struct {
	Repository *repositories.MultipleRepository
}

func (t *MultipleTask) Start() error {
	return t.Repository.Start()
}

func (t *MultipleTask) Stop() error {
	return t.Repository.Stop()
}

func (t *MultipleTask) Clean() error {
	return t.Repository.Clean()
}
