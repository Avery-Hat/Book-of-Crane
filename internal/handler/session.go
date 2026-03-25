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

type SessionHandler struct {
	store *store.SessionStore
}

func NewSessionHandler(s *store.SessionStore) *SessionHandler {
	return &SessionHandler{store: s}
}

func (h *SessionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{sessionID}", h.Get)
	r.Put("/{sessionID}", h.Update)
	r.Delete("/{sessionID}", h.Delete)

	r.Get("/{sessionID}/recap", h.Recap)

	// Entity links nested under a session
	r.Get("/{sessionID}/npcs", h.ListNPCs)
	r.Post("/{sessionID}/npcs", h.LinkNPC)
	r.Delete("/{sessionID}/npcs/{npcID}", h.UnlinkNPC)

	r.Get("/{sessionID}/locations", h.ListLocations)
	r.Post("/{sessionID}/locations", h.LinkLocation)
	r.Delete("/{sessionID}/locations/{locationID}", h.UnlinkLocation)

	r.Get("/{sessionID}/items", h.ListItems)
	r.Post("/{sessionID}/items", h.LinkItem)
	r.Delete("/{sessionID}/items/{itemID}", h.UnlinkItem)

	return r
}

func sessionID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "sessionID"))
}

// --- CRUD ---

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	sessions, err := h.store.List(r.Context(), cid)
	if err != nil {
		log.Printf("ERROR listing sessions: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	if sessions == nil {
		sessions = []model.Session{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	session, err := h.store.GetByID(r.Context(), cid, sid)
	if err != nil {
		log.Printf("ERROR getting session %d: %v", sid, err)
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	var req model.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SessionNumber < 1 {
		writeError(w, http.StatusBadRequest, "session_number must be a positive integer")
		return
	}
	session, err := h.store.Create(r.Context(), cid, req)
	if err != nil {
		log.Printf("ERROR creating session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	var req model.UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := h.store.Update(r.Context(), cid, sid, req)
	if err != nil {
		log.Printf("ERROR updating session %d: %v", sid, err)
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	if err := h.store.Delete(r.Context(), cid, sid); err != nil {
		log.Printf("ERROR deleting session %d: %v", sid, err)
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Recap ---

func (h *SessionHandler) Recap(w http.ResponseWriter, r *http.Request) {
	cid, err := campaignID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	recap, err := h.store.Recap(r.Context(), cid, sid)
	if err != nil {
		log.Printf("ERROR getting recap for session %d: %v", sid, err)
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, recap)
}

// --- Entity links ---

func (h *SessionHandler) ListNPCs(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	links, err := h.store.ListNPCs(r.Context(), sid)
	if err != nil {
		log.Printf("ERROR listing session npcs: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list session npcs")
		return
	}
	if links == nil {
		links = []model.SessionNPCLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *SessionHandler) LinkNPC(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	var req model.LinkSessionNPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.NPCID == 0 {
		writeError(w, http.StatusBadRequest, "npc_id is required")
		return
	}
	if err := h.store.LinkNPC(r.Context(), sid, req); err != nil {
		log.Printf("ERROR linking npc to session %d: %v", sid, err)
		writeError(w, http.StatusInternalServerError, "failed to link npc to session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) UnlinkNPC(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	if err := h.store.UnlinkNPC(r.Context(), sid, nid); err != nil {
		log.Printf("ERROR unlinking npc from session %d: %v", sid, err)
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	links, err := h.store.ListLocations(r.Context(), sid)
	if err != nil {
		log.Printf("ERROR listing session locations: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list session locations")
		return
	}
	if links == nil {
		links = []model.SessionLocationLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *SessionHandler) LinkLocation(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	var req model.LinkSessionLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.LocationID == 0 {
		writeError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	if err := h.store.LinkLocation(r.Context(), sid, req); err != nil {
		log.Printf("ERROR linking location to session %d: %v", sid, err)
		writeError(w, http.StatusInternalServerError, "failed to link location to session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) UnlinkLocation(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	lid, err := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid location id")
		return
	}
	if err := h.store.UnlinkLocation(r.Context(), sid, lid); err != nil {
		log.Printf("ERROR unlinking location from session %d: %v", sid, err)
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	links, err := h.store.ListItems(r.Context(), sid)
	if err != nil {
		log.Printf("ERROR listing session items: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list session items")
		return
	}
	if links == nil {
		links = []model.SessionItemLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *SessionHandler) LinkItem(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	var req model.LinkSessionItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ItemID == 0 {
		writeError(w, http.StatusBadRequest, "item_id is required")
		return
	}
	if err := h.store.LinkItem(r.Context(), sid, req); err != nil {
		log.Printf("ERROR linking item to session %d: %v", sid, err)
		writeError(w, http.StatusInternalServerError, "failed to link item to session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) UnlinkItem(w http.ResponseWriter, r *http.Request) {
	sid, err := sessionID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	iid, err := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	if err := h.store.UnlinkItem(r.Context(), sid, iid); err != nil {
		log.Printf("ERROR unlinking item from session %d: %v", sid, err)
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
