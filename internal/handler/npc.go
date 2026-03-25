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

type NPCHandler struct {
	store *store.NPCStore
}

func NewNPCHandler(s *store.NPCStore) *NPCHandler {
	return &NPCHandler{store: s}
}

func (h *NPCHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{npcID}", h.Get)
	r.Put("/{npcID}", h.Update)
	r.Delete("/{npcID}", h.Delete)
	r.Get("/{npcID}/detail", h.Detail)
	return r
}

func campaignID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "campaignID"))
}

func (h *NPCHandler) List(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	npcs, err := h.store.List(r.Context(), cid)
	if err != nil {
		log.Printf("ERROR listing npcs: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list npcs")
		return
	}
	if npcs == nil {
		npcs = []model.NPC{}
	}
	writeJSON(w, http.StatusOK, npcs)
}

func (h *NPCHandler) Get(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	npc, err := h.store.GetByID(r.Context(), cid, nid)
	if err != nil {
		log.Printf("ERROR getting npc %d: %v", nid, err)
		writeError(w, http.StatusNotFound, "npc not found")
		return
	}
	writeJSON(w, http.StatusOK, npc)
}

func (h *NPCHandler) Create(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	var req model.CreateNPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	npc, err := h.store.Create(r.Context(), cid, req)
	if err != nil {
		log.Printf("ERROR creating npc: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create npc")
		return
	}
	writeJSON(w, http.StatusCreated, npc)
}

func (h *NPCHandler) Update(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	var req model.UpdateNPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	npc, err := h.store.Update(r.Context(), cid, nid, req)
	if err != nil {
		log.Printf("ERROR updating npc %d: %v", nid, err)
		writeError(w, http.StatusNotFound, "npc not found")
		return
	}
	writeJSON(w, http.StatusOK, npc)
}

func (h *NPCHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	if err := h.store.Delete(r.Context(), cid, nid); err != nil {
		log.Printf("ERROR deleting npc %d: %v", nid, err)
		writeError(w, http.StatusNotFound, "npc not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NPCHandler) Detail(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	detail, err := h.store.Detail(r.Context(), cid, nid)
	if err != nil {
		log.Printf("ERROR getting npc detail %d: %v", nid, err)
		writeError(w, http.StatusNotFound, "npc not found")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}
