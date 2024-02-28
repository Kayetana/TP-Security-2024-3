package api_delivery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"proxy/internal/api"
	"proxy/internal/proxy"
	"proxy/internal/utils"
	"strconv"
)

type Handler struct {
	repo  api.Repository
	proxy proxy.HandlerProxy
}

func NewHandler(repo api.Repository, proxy proxy.HandlerProxy) *Handler {
	return &Handler{repo: repo, proxy: proxy}
}

func (h *Handler) GetAllRequests(w http.ResponseWriter, r *http.Request) {
	log.Println("entered handler GetAllRequests")

	requests, err := h.repo.GetAllRequests()
	if err != nil {
		log.Println("error getting requests:", err)
		h.writeJSON(w, http.StatusInternalServerError, api.Error{Error: "failed to get requests"})
		return
	}
	log.Println("got requests")

	h.writeJSON(w, http.StatusOK, requests)
}

func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("entered handler GetRequest")

	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		log.Println("error getting id:", err)
		h.writeJSON(w, http.StatusBadRequest, api.Error{Error: "correct id required"})
		return
	}
	log.Println("got id", id)

	request, err := h.repo.GetRequestById(id)
	if err != nil {
		h.writeJSON(w, http.StatusNotFound, api.Error{Error: "no such request"})
		return
	}
	h.writeJSON(w, http.StatusOK, request)
}

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("entered handler RepeatRequest")

	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		log.Println("error getting id:", err)
		h.writeJSON(w, http.StatusBadRequest, api.Error{Error: "correct id required"})
		return
	}
	log.Println("got id", id)

	request, err := h.repo.GetRequestById(id)
	if err != nil {
		log.Println("error getting request:", err)
		h.writeJSON(w, http.StatusNotFound, api.Error{Error: "no such request"})
		return
	}
	log.Println("got request by id")

	response, err := h.proxy.SendRequest(&request)
	if err != nil {
		log.Println("error sending request:", err)
		h.writeJSON(w, http.StatusInternalServerError, api.Error{})
		return
	}
	defer response.Body.Close()

	w.WriteHeader(response.StatusCode)
	log.Println("got response with status code", response.StatusCode)
	log.Println("headers:", response.Header)
	utils.CopyHeaders(w.Header(), response.Header)

	if _, err = io.Copy(w, response.Body); err != nil {
		log.Println("error writing body:", err)
		h.writeJSON(w, http.StatusInternalServerError, api.Error{})
	}
}
