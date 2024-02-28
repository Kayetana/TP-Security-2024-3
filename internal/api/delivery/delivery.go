package api_delivery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"proxy/internal/api"
	"strconv"
)

type Handler struct {
	repo api.Repository
}

func NewHandler(repo api.Repository) *Handler {
	return &Handler{repo: repo}
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
