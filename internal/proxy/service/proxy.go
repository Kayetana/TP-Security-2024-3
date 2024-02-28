package proxy_service

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
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

func (p *Proxy) Start(port string) error {
	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				p.handleHTTPS(w, r)
			} else {
				p.handleHTTP(w, r)
			}
		}),
	}
	return server.ListenAndServe()
}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("got http request Method:", r.Method, "Host:", r.Host)

	parsedReq, err := p.parseRequest(r)
	if err != nil {
		log.Println("request parsing error:", err)
	} else {
		err = p.repo.SaveRequest(parsedReq)
		if err != nil {
			log.Println("saving request to db error:", err)
		}
	}

	r.Header.Del("Proxy-Connection")
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	log.Println("response status code:", resp.StatusCode)

	parsedResp, err := p.parseResponse(resp)
	if err != nil {
		log.Println("response parsing error:", err)
	} else {
		parsedResp.RequestId = parsedReq.Id
		err = p.repo.SaveResponse(parsedResp)
		if err != nil {
			log.Println("saving response to db error:", err)
		}
	}

	w.WriteHeader(resp.StatusCode)
	utils.CopyHeaders(w.Header(), resp.Header)

	if _, err = io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *Proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Println("got https request", "Host:", r.Host)

	clientConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()
	log.Println("got underlying TCP connection")

	if _, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
		log.Println("error writing response:", err)
		return
	}

	config, err := p.getConfig(strings.Split(r.Host, ":")[0])
	if err != nil {
		log.Println("certificate generation error:", err)
		return
	}

	tlsClientConn := tls.Server(clientConn, &config)
	defer tlsClientConn.Close()

	if err = tlsClientConn.Handshake(); err != nil {
		log.Println("error establishing connection to client:", err)
		return
	}
	log.Println("TLS connection with client established")

	tlsServerConn, err := tls.Dial("tcp", r.Host, &config)
	if err != nil {
		log.Println("error establishing connection to server", err)
		return
	}
	defer tlsServerConn.Close()
	log.Println("TLS connection with server established")

	req, err := p.passRequest(tlsClientConn, tlsServerConn)
	if err != nil {
		log.Println("request passing error:", err)
		return
	}
	log.Println("request has been passed")

	parsedReq, err := p.parseRequest(req)
	if err != nil {
		log.Println(err)
	}
	if err = p.repo.SaveRequest(parsedReq); err != nil {
		log.Println("saving request to db error:", err)
	}

	resp, err := p.passResponse(tlsClientConn, tlsServerConn, req)
	if err != nil {
		log.Println("response passing error", err)
		return
	}
	log.Println("response has been passed")

	parsedResp, err := p.parseResponse(resp)
	if err != nil {
		log.Println(err)
	}
	parsedResp.RequestId = parsedReq.Id
	if err = p.repo.SaveResponse(parsedResp); err != nil {
		log.Println("saving response to db error:", err)
	}
}

func (p *Proxy) getConfig(host string) (tls.Config, error) {
	if err := exec.Command(p.certs["pathToGenCert"], host, strconv.Itoa(rand.Int())).Run(); err != nil {
		return tls.Config{}, err
	}

	generatedCert, err := tls.LoadX509KeyPair(p.certs["pathToCertFile"], p.certs["pathToKeyFile"])
	if err != nil {
		return tls.Config{}, err
	}

	return tls.Config{
		Certificates: []tls.Certificate{generatedCert},
		ServerName:   host,
	}, nil
}

func (p *Proxy) passRequest(clientConn, serverConn *tls.Conn) (*http.Request, error) {
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

func (p *Proxy) passResponse(clientConn, serverConn *tls.Conn, request *http.Request) (*http.Response, error) {
	response, err := http.ReadResponse(bufio.NewReader(serverConn), request)
	if err != nil {
		return &http.Response{}, err
	}
	rawResponse, err := httputil.DumpResponse(response, true)
	if err != nil {
		return &http.Response{}, err
	}
	_, err = clientConn.Write(rawResponse)
	return response, err
}

func (p *Proxy) parseRequest(req *http.Request) (*proxy.Request, error) {
	parsedReq := &proxy.Request{
		Method:      req.Method,
		Path:        fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path),
		ContentType: req.Header.Get("Content-Type"),
	}

	queryParams, _ := json.Marshal(req.URL.Query())
	queryParamsRaw := json.RawMessage(queryParams)
	parsedReq.QueryParams = &queryParamsRaw

	headers, _ := json.Marshal(req.Header)
	headersRaw := json.RawMessage(headers)
	parsedReq.Headers = &headersRaw

	cookies, _ := json.Marshal(req.Cookies())
	cookiesRaw := json.RawMessage(cookies)
	parsedReq.Cookies = &cookiesRaw

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return parsedReq, errors.New("error parsing request body")
	}
	parsedReq.Body = string(body)
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return parsedReq, nil
}

func (p *Proxy) parseResponse(resp *http.Response) (*proxy.Response, error) {
	parsedResp := &proxy.Response{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
	}

	headers, _ := json.Marshal(resp.Header)
	headersRaw := json.RawMessage(headers)
	parsedResp.Headers = &headersRaw

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return parsedResp, errors.New("error parsing response body")
	}
	parsedResp.Body = string(body)
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return parsedResp, nil
}
