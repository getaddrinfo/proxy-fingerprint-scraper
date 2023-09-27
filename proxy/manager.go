package proxy

import (
	"errors"
)

var ErrNoneMatch = errors.New("no proxies are usable at this current moment")

type Manager interface {
	Init() error
	IPs() []string
	Add(ip string) error
	Get() (Proxy, error)
}
