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

type FactionHandler struct {
	store *store.FactionStore
}

func NewFactionHandler(s *store.FactionStore) *FactionHandler {
	return &FactionHandler{store: s}
}

func (h *FactionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{factionID}", h.Get)
	r.Put("/{factionID}", h.Update)
	r.Delete("/{factionID}", h.Delete)
	return r
}

func (h *FactionHandler) List(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	factions, err := h.store.List(r.Context(), cid)
	if err != nil {
		log.Printf("ERROR listing factions: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list factions")
		return
	}
	if factions == nil {
		factions = []model.Faction{}
	}
	writeJSON(w, http.StatusOK, factions)
}

func (h *FactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	fid, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid faction id")
		return
	}
	faction, err := h.store.GetByID(r.Context(), cid, fid)
	if err != nil {
		log.Printf("ERROR getting faction %d: %v", fid, err)
		writeError(w, http.StatusNotFound, "faction not found")
		return
	}
	writeJSON(w, http.StatusOK, faction)
}

func (h *FactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	var req model.CreateFactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	faction, err := h.store.Create(r.Context(), cid, req)
	if err != nil {
		log.Printf("ERROR creating faction: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create faction")
		return
	}
	writeJSON(w, http.StatusCreated, faction)
}

func (h *FactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	fid, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid faction id")
		return
	}
	var req model.UpdateFactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	faction, err := h.store.Update(r.Context(), cid, fid, req)
	if err != nil {
		log.Printf("ERROR updating faction %d: %v", fid, err)
		writeError(w, http.StatusNotFound, "faction not found")
		return
	}
	writeJSON(w, http.StatusOK, faction)
}

func (h *FactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	fid, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid faction id")
		return
	}
	if err := h.store.Delete(r.Context(), cid, fid); err != nil {
		log.Printf("ERROR deleting faction %d: %v", fid, err)
		writeError(w, http.StatusNotFound, "faction not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
