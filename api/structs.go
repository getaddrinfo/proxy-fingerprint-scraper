package api

type StatsResponse struct {
	Count   uint64   `json:"count"`
	Proxies []string `json:"proxies"`
}

type GetFingerprintResponse struct {
	ID          uint64 `json:"id"`
	Fingerprint string `json:"fingerprint"`
	ProxyIP     string `json:"proxy_ip"`
}

type ErrorResponse struct {
	Error string  `json:"message"`
	Code  *string `json:"code,omitempty"`
}
