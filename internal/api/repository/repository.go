package api_repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"proxy/internal/api"
	"proxy/internal/proxy"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) api.Repository {
	return &Postgres{Pool: pool}
}

func (db *Postgres) GetAllRequests() ([]proxy.Request, error) {
	requests := make([]proxy.Request, 0)

	const query = `select id, method, path, query_params, headers, cookies, content_type, body from request`
	rows, err := db.Pool.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r proxy.Request

		err = rows.Scan(&r.Id, &r.Method, &r.Path, &r.QueryParams, &r.Headers, &r.Cookies, &r.ContentType, &r.Body)
		if err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}

	return requests, nil
}

func (db *Postgres) GetRequestById(id uint64) (proxy.Request, error) {
	var r proxy.Request
	const query = `select id, method, path, query_params, headers, cookies, content_type, body from request where id = $1`

	row := db.Pool.QueryRow(context.TODO(), query, id)
	err := row.Scan(&r.Id, &r.Method, &r.Path, &r.QueryParams, &r.Headers, &r.Cookies, &r.ContentType, &r.Body)
	if err != nil {
		return proxy.Request{}, err
	}
	return r, nil
}
