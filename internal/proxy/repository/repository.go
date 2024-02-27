package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"proxy/internal/proxy"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) proxy.Repository {
	return &Postgres{Pool: pool}
}

func (db *Postgres) SaveRequest(req *proxy.Request) error {
	const query = `insert into request (method, path, query_params, headers, cookies, content_type, body) values ($1, $2, $3, $4, $5, $6, $7) returning id`
	var requestId uint64

	row := db.Pool.QueryRow(context.TODO(), query, req.Method, req.Path, req.QueryParams, req.Headers, req.Cookies, req.ContentType, req.Body)
	err := row.Scan(&requestId)
	req.Id = requestId

	return err
}

func (db *Postgres) SaveResponse(resp *proxy.Response) error {
	const query = `insert into response (request_id, status_code, headers, content_type, body) values ($1, $2, $3, $4, $5)`
	_, err := db.Pool.Exec(context.TODO(), query, resp.RequestId, resp.StatusCode, resp.Headers, resp.ContentType, resp.Body)
	return err
}
