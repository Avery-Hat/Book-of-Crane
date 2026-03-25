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

type ItemHandler struct {
	store *store.ItemStore
}

func NewItemHandler(s *store.ItemStore) *ItemHandler {
	return &ItemHandler{store: s}
}

func (h *ItemHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{itemID}", h.Get)
	r.Put("/{itemID}", h.Update)
	r.Delete("/{itemID}", h.Delete)
	return r
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	items, err := h.store.List(r.Context(), cid)
	if err != nil {
		log.Printf("ERROR listing items: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list items")
		return
	}
	if items == nil {
		items = []model.Item{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	iid, err := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	item, err := h.store.GetByID(r.Context(), cid, iid)
	if err != nil {
		log.Printf("ERROR getting item %d: %v", iid, err)
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	var req model.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	item, err := h.store.Create(r.Context(), cid, req)
	if err != nil {
		log.Printf("ERROR creating item: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create item")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	iid, err := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	var req model.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name != nil && *req.Name == "" {
		writeError(w, http.StatusBadRequest, "name cannot be empty")
		return
	}
	item, err := h.store.Update(r.Context(), cid, iid, req)
	if err != nil {
		log.Printf("ERROR updating item %d: %v", iid, err)
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	iid, err := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	if err := h.store.Delete(r.Context(), cid, iid); err != nil {
		log.Printf("ERROR deleting item %d: %v", iid, err)
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
