package main

import (
	"io"
	"log"
	"net/http"
)

const proxyPort = "8080"

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("entered handleHTTP\tMethod:", r.Method, "Host:", r.Host)

	r.Header.Del("Proxy-Connection")
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	log.Println("response has been received. StatusCode:", resp.StatusCode)

	w.WriteHeader(resp.StatusCode)
	copyHeaders(w.Header(), resp.Header)

	if _, err = io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func copyHeaders(dst, src http.Header) {
	for name, values := range src {
		for _, value := range values {
			dst.Add(name, value)
		}
	}
}

func main() {
	proxy := &http.Server{
		Addr: ":" + proxyPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleHTTP(w, r)
		}),
	}

	log.Println("starting proxy on", proxy.Addr)
	log.Fatal(proxy.ListenAndServe())
}
