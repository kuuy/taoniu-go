package dice

import repositories "taoniu.local/gamblings/repositories/wolf/dice"

type BetTask struct {
	Repository *repositories.BetRepository
}

func (t *BetTask) Start() {
	t.Repository.Start()
}

func (t *BetTask) Stop() {
	t.Repository.Stop()
}
