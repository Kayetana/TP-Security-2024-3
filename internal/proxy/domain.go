package proxy

import "net/http"

type Repository interface {
	SaveRequest(r *http.Request) (int, error)
}
