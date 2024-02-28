package api

import (
	"encoding/json"
)

type Repository interface {
	GetAllRequests() ([]Request, error)
	GetRequestById(id uint64) (Request, error)
}

type Request struct {
	Id          uint64           `json:"id"`
	Method      string           `json:"method"`
	Path        string           `json:"path"`
	QueryParams *json.RawMessage `json:"query_params"`
	Headers     *json.RawMessage `json:"headers"`
	Cookies     *json.RawMessage `json:"cookies"`
	ContentType string           `json:"content_type"`
	Body        string           `json:"body"`
}

type Error struct {
	Error string `json:"error"`
}
