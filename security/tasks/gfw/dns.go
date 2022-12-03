package gfw

import repositories "taoniu.local/security/repositories/gfw"

type DnsTask struct {
	Repository *repositories.DnsRepository
}

func (t *DnsTask) Flush() error {
	return t.Repository.Flush()
}

func (t *DnsTask) Cache() error {
	return t.Repository.Cache()
}
