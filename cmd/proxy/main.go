package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	api_delivery "proxy/internal/api/delivery"
	api_repository "proxy/internal/api/repository"
	proxy_repository "proxy/internal/proxy/repository"
	proxy_service "proxy/internal/proxy/service"
)

const (
	apiPort              = "8000"
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

	proxyRepo := proxy_repository.NewPostgres(connPool)
	proxy := proxy_service.NewProxy(proxyRepo, pathToGenCert, pathToCertFile, pathToKeyFile)

	log.Println("starting proxy server on", proxyPort)
	go func() {
		log.Fatal(proxy.Start(proxyPort))
	}()

	apiRepo := api_repository.NewPostgres(connPool)
	handler := api_delivery.NewHandler(apiRepo, proxy)

	router := mux.NewRouter()
	router.HandleFunc("/requests", handler.GetAllRequests).Methods(http.MethodGet)
	router.HandleFunc("/requests/{id}", handler.GetRequest).Methods(http.MethodGet)
	router.HandleFunc("/repeat/{id}", handler.RepeatRequest).Methods(http.MethodGet)
	router.HandleFunc("/scan/{id}", handler.ScanRequest).Methods(http.MethodGet)

	log.Println("starting api server on", apiPort)
	log.Fatal(http.ListenAndServe(":"+apiPort, router))
}
