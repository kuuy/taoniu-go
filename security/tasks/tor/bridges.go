package tor

import repositories "taoniu.local/security/repositories/tor"

type BridgesTask struct {
	Repository *repositories.BridgesRepository
}

func (t *BridgesTask) Flush() error {
	return t.Repository.Flush()
}

func (t *BridgesTask) Rescue() error {
	return t.Repository.Rescue()
}
