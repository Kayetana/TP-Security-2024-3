package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"proxy/internal/proxy"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) proxy.Repository {
	return &Postgres{Pool: pool}
}

func (db *Postgres) SaveRequest(r *http.Request) (int, error) {
	queryString := "insert into request (method, body) values ($1, $2) returning id"
	row := db.Pool.QueryRow(context.TODO(), queryString, r.Method, "")

	var requestId int
	err := row.Scan(&requestId)

	return requestId, err
}
