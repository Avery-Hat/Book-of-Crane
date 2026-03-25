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

type CampaignHandler struct {
	store *store.CampaignStore
}

func NewCampaignHandler(s *store.CampaignStore) *CampaignHandler {
	return &CampaignHandler{store: s}
}

func (h *CampaignHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.store.List(r.Context())
	if err != nil {
		log.Printf("ERROR listing campaigns: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list campaigns")
		return
	}
	if campaigns == nil {
		campaigns = []model.Campaign{}
	}
	writeJSON(w, http.StatusOK, campaigns)
}

func (h *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	campaign, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("ERROR getting campaign %d: %v", id, err)
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	campaign, err := h.store.Create(r.Context(), req)
	if err != nil {
		log.Printf("ERROR creating campaign: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create campaign")
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	var req model.UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name != nil && *req.Name == "" {
		writeError(w, http.StatusBadRequest, "name cannot be empty")
		return
	}

	campaign, err := h.store.Update(r.Context(), id, req)
	if err != nil {
		log.Printf("ERROR updating campaign %d: %v", id, err)
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	if err := h.store.Delete(r.Context(), id); err != nil {
		log.Printf("ERROR deleting campaign %d: %v", id, err)
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
