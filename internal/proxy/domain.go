package proxy

import (
	"encoding/json"
	"net/http"
)

type Repository interface {
	SaveRequest(req *Request) error
	SaveResponse(resp *Response) error
}

type HandlerProxy interface {
	SendRequest(req *Request) (*http.Response, error)
	CheckForInjection(req *Request) (string, error)
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

type Response struct {
	Id          uint64
	RequestId   uint64
	StatusCode  int
	Headers     *json.RawMessage
	ContentType string
	Body        string
}
