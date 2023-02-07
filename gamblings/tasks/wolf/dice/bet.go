package dice

import repositories "taoniu.local/gamblings/repositories/wolf/dice"

type BetTask struct {
	Repository *repositories.BetRepository
}

func (t *BetTask) Start(strategy string) {
	t.Repository.Start(strategy)
}

func (t *BetTask) Stop(strategy string) {
	t.Repository.Stop(strategy)
}
