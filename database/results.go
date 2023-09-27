package database

import "github.com/getaddrinfo/proxy-fingerprint-scraper/common"

type GetFingerprintResult struct {
	ID          uint64
	Fingerprint string
	ProxyIP     string
}

type GetAuthResult struct {
	Valid       bool
	UserId      uint64
	Permissions common.Permission
}

type GetUserResult struct {
	UserId      uint64
	Permissions common.Permission
	Token       string
}
