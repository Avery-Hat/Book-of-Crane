package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/Avery-Hat/Book-of-Crane/internal/store"
	"github.com/go-chi/chi/v5"
)

type LocationHandler struct {
	store *store.LocationStore
}

func NewLocationHandler(s *store.LocationStore) *LocationHandler {
	return &LocationHandler{store: s}
}

func (h *LocationHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{locationID}", h.Get)
	r.Put("/{locationID}", h.Update)
	r.Delete("/{locationID}", h.Delete)
	return r
}

func (h *LocationHandler) List(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	locations, err := h.store.List(r.Context(), cid)
	if err != nil {
		log.Printf("ERROR listing locations: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list locations")
		return
	}
	if locations == nil {
		locations = []model.Location{}
	}
	writeJSON(w, http.StatusOK, locations)
}

func (h *LocationHandler) Get(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	lid, err := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid location id")
		return
	}
	location, err := h.store.GetByID(r.Context(), cid, lid)
	if err != nil {
		log.Printf("ERROR getting location %d: %v", lid, err)
		writeError(w, http.StatusNotFound, "location not found")
		return
	}
	writeJSON(w, http.StatusOK, location)
}

func (h *LocationHandler) Create(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	var req model.CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	location, err := h.store.Create(r.Context(), cid, req)
	if err != nil {
		log.Printf("ERROR creating location: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create location")
		return
	}
	writeJSON(w, http.StatusCreated, location)
}

func (h *LocationHandler) Update(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	lid, err := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid location id")
		return
	}
	var req model.UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	location, err := h.store.Update(r.Context(), cid, lid, req)
	if err != nil {
		log.Printf("ERROR updating location %d: %v", lid, err)
		writeError(w, http.StatusNotFound, "location not found")
		return
	}
	writeJSON(w, http.StatusOK, location)
}

func (h *LocationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	lid, err := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid location id")
		return
	}
	if err := h.store.Delete(r.Context(), cid, lid); err != nil {
		log.Printf("ERROR deleting location %d: %v", lid, err)
		writeError(w, http.StatusNotFound, "location not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
