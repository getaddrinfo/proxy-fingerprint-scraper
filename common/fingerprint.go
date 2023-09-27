package common

type FingerprintResult struct {
	Fingerprint string
	ProxyIP     string
}

type FingerprintResultChannel chan FingerprintResult
