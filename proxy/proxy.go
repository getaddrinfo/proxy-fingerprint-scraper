package proxy

import (
	"net/http"
)

const GenerationPerTwoMinuteLimit = 3
const FailedRequestImpliesDownLimit = 2

type Proxy interface {
	IP() string
	Client() *http.Client
	Do(req *http.Request) (*http.Response, error)
}
