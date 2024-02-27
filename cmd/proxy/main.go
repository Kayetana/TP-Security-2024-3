package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"proxy/internal/proxy/repository"
	proxy "proxy/internal/proxy/service"
)

const (
	proxyPort            = "8080"
	pathToGenCert        = "/certs/gen_cert.sh"
	pathToCertFile       = "/certs/cert.crt"
	pathToKeyFile        = "/certs/cert.key"
	EnvPostgresQueryName = "DATABASE_URL"
)

func main() {
	connPool, err := pgxpool.New(context.Background(), os.Getenv(EnvPostgresQueryName))
	if err != nil {
		log.Fatal("error connecting to db:", err)
	}
	log.Println("connected to db")

	repo := repository.NewPostgres(connPool)
	prx := proxy.NewProxy(repo, pathToGenCert, pathToCertFile, pathToKeyFile)
	log.Fatal(prx.Start(proxyPort))
}
