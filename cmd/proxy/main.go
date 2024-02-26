package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"strconv"
	"strings"
)

const (
	proxyPort      = "8080"
	pathToGenCert  = "/certs/gen_cert.sh"
	pathToCertFile = "/certs/cert.crt"
	pathToKeyFile  = "/certs/cert.key"
)

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

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Println("entered handleHTTPS\tMethod:", r.Method, "Host:", r.Host)

	clientConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()
	log.Println("got underlying TCP connection")

	if _, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
		log.Println("error", err)
		return
	}

	config, err := getConfig(strings.Split(r.Host, ":")[0])
	if err != nil {
		log.Println("error", err)
		return
	}

	tlsClientConn := tls.Server(clientConn, &config)
	defer tlsClientConn.Close()

	if err = tlsClientConn.Handshake(); err != nil {
		log.Println("error", err)
		return
	}
	log.Println("TLS connection with client established")

	tlsServerConn, err := tls.Dial("tcp", r.Host, &config)
	if err != nil {
		log.Println("error", err)
		return
	}
	defer tlsServerConn.Close()
	log.Println("TLS connection with server established")

	req, err := passRequest(tlsClientConn, tlsServerConn)
	if err != nil {
		log.Println("error", err)
		return
	}
	log.Println("request was sent")
	err = passResponse(tlsClientConn, tlsServerConn, req)
	if err != nil {
		log.Println("error", err)
		return
	}
	log.Println("response was sent")
}

func getConfig(host string) (tls.Config, error) {
	if err := exec.Command(pathToGenCert, host, strconv.Itoa(rand.Int())).Run(); err != nil {
		log.Println("error generating file", pathToCertFile)
		return tls.Config{}, err
	}

	cert, err := tls.LoadX509KeyPair(pathToCertFile, pathToKeyFile)
	if err != nil {
		log.Println("error getting certificate", err)
		return tls.Config{}, err
	}

	return tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   host,
	}, nil
}

func passRequest(clientConn, serverConn *tls.Conn) (*http.Request, error) {
	request, err := http.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		return &http.Request{}, err
	}
	rawRequest, err := httputil.DumpRequest(request, true)
	if err != nil {
		return &http.Request{}, err
	}
	_, err = serverConn.Write(rawRequest)
	return request, err
}

func passResponse(clientConn, serverConn *tls.Conn, request *http.Request) error {
	response, err := http.ReadResponse(bufio.NewReader(serverConn), request)
	if err != nil {
		return err
	}
	rawResponse, err := httputil.DumpResponse(response, true)
	if err != nil {
		return err
	}
	_, err = clientConn.Write(rawResponse)
	return err
}

func main() {
	proxy := &http.Server{
		Addr: ":" + proxyPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleHTTPS(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
	}

	log.Println("starting proxy on", proxy.Addr)
	log.Fatal(proxy.ListenAndServe())
}
