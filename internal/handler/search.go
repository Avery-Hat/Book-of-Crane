package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/Avery-Hat/Book-of-Crane/internal/store"
	"github.com/go-chi/chi/v5"
)

type SearchHandler struct {
	store *store.SearchStore
}

func NewSearchHandler(s *store.SearchStore) *SearchHandler {
	return &SearchHandler{store: s}
}

func (h *SearchHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Search)
	return r
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeError(w, http.StatusBadRequest, "q is required")
		return
	}

	results, err := h.store.Search(r.Context(), cid, q)
	if err != nil {
		log.Printf("ERROR searching campaign %d: %v", cid, err)
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	writeJSON(w, http.StatusOK, results)
}
