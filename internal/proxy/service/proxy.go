package proxy_service

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"proxy/internal/proxy"
	"proxy/internal/utils"
	"strconv"
	"strings"
)

type Proxy struct {
	repo  proxy.Repository
	certs map[string]string
}

func NewProxy(repo proxy.Repository, pathToGenCert, pathToCertFile, pathToKeyFile string) *Proxy {
	certs := map[string]string{
		"pathToGenCert":  pathToGenCert,
		"pathToCertFile": pathToCertFile,
		"pathToKeyFile":  pathToKeyFile,
	}

	return &Proxy{
		repo:  repo,
		certs: certs,
	}
}

func (proxy *Proxy) Start(port string) error {
	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				proxy.handleHTTPS(w, r)
			} else {
				proxy.handleHTTP(w, r)
			}
		}),
	}
	log.Println("starting server on", server.Addr)
	return server.ListenAndServe()
}

func (proxy *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("entered handleHTTP\tMethod:", r.Method, "Host:", r.Host)

	r.Header.Del("Proxy-Connection")

	requestId, err := proxy.repo.SaveRequest(r)
	if err != nil {
		log.Println("error saving to db")
	}
	log.Println("requestId", requestId)

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	log.Println("response has been received. StatusCode:", resp.StatusCode)

	w.WriteHeader(resp.StatusCode)
	utils.CopyHeaders(w.Header(), resp.Header)

	if _, err = io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (proxy *Proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {
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

	config, err := proxy.getConfig(strings.Split(r.Host, ":")[0])
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

	req, err := proxy.passRequest(tlsClientConn, tlsServerConn)
	if err != nil {
		log.Println("error", err)
		return
	}
	log.Println("request was sent")
	err = proxy.passResponse(tlsClientConn, tlsServerConn, req)
	if err != nil {
		log.Println("error", err)
		return
	}
	log.Println("response was sent")
}

func (proxy *Proxy) getConfig(host string) (tls.Config, error) {
	if err := exec.Command(proxy.certs["pathToGenCert"], host, strconv.Itoa(rand.Int())).Run(); err != nil {
		log.Println("certificate generating error")
		return tls.Config{}, err
	}

	generatedCert, err := tls.LoadX509KeyPair(proxy.certs["pathToCertFile"], proxy.certs["pathToKeyFile"])
	if err != nil {
		log.Println("error getting certificate", err)
		return tls.Config{}, err
	}

	return tls.Config{
		Certificates: []tls.Certificate{generatedCert},
		ServerName:   host,
	}, nil
}

func (proxy *Proxy) passRequest(clientConn, serverConn *tls.Conn) (*http.Request, error) {
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

func (proxy *Proxy) passResponse(clientConn, serverConn *tls.Conn, request *http.Request) error {
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
