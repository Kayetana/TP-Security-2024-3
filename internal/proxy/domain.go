package proxy

import (
	"encoding/json"
)

type Repository interface {
	SaveRequest(req *Request) error
	SaveResponse(resp *Response) error
}

type Request struct {
	Id          uint64
	Method      string
	Path        string
	QueryParams *json.RawMessage
	Headers     *json.RawMessage
	Cookies     *json.RawMessage
	ContentType string
	Body        string
}

type Response struct {
	Id          uint64
	RequestId   uint64
	StatusCode  int
	Headers     *json.RawMessage
	ContentType string
	Body        string
}
