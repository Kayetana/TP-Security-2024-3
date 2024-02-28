package api

import (
	"proxy/internal/proxy"
)

type Repository interface {
	GetAllRequests() ([]proxy.Request, error)
	GetRequestById(id uint64) (proxy.Request, error)
}

type Error struct {
	Error string `json:"error"`
}

type ScanningResult struct {
	Result string `json:"scanning_result"`
}
