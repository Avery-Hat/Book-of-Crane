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

type RelationshipHandler struct {
	store *store.RelationshipStore
}

func NewRelationshipHandler(s *store.RelationshipStore) *RelationshipHandler {
	return &RelationshipHandler{store: s}
}

func (h *RelationshipHandler) NPCFactionRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListNPCFactions)
	r.Post("/", h.LinkNPCFaction)
	r.Delete("/{factionID}", h.UnlinkNPCFaction)
	return r
}

func (h *RelationshipHandler) NPCLocationRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListNPCLocations)
	r.Post("/", h.LinkNPCLocation)
	r.Delete("/{locationID}", h.UnlinkNPCLocation)
	return r
}

func (h *RelationshipHandler) FactionLocationRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListFactionLocations)
	r.Post("/", h.LinkFactionLocation)
	r.Delete("/{locationID}", h.UnlinkFactionLocation)
	return r
}

func (h *RelationshipHandler) NPCRelationshipRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListNPCRelationships)
	r.Post("/", h.CreateNPCRelationship)
	r.Delete("/{otherNPCID}", h.DeleteNPCRelationship)
	return r
}

// --- NPC <-> Faction ---

func (h *RelationshipHandler) ListNPCFactions(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	links, err := h.store.ListNPCFactions(r.Context(), nid)
	if err != nil {
		log.Printf("ERROR listing npc factions: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list npc factions")
		return
	}
	if links == nil {
		links = []model.NPCFactionLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *RelationshipHandler) LinkNPCFaction(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	var req model.LinkNPCFactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.FactionID == 0 {
		writeError(w, http.StatusBadRequest, "faction_id is required")
		return
	}
	if err := h.store.LinkNPCFaction(r.Context(), nid, req); err != nil {
		log.Printf("ERROR linking npc %d to faction: %v", nid, err)
		writeError(w, http.StatusInternalServerError, "failed to link npc to faction")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RelationshipHandler) UnlinkNPCFaction(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	fid, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid faction id")
		return
	}
	if err := h.store.UnlinkNPCFaction(r.Context(), nid, fid); err != nil {
		log.Printf("ERROR unlinking npc %d from faction %d: %v", nid, fid, err)
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- NPC <-> Location ---

func (h *RelationshipHandler) ListNPCLocations(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	links, err := h.store.ListNPCLocations(r.Context(), nid)
	if err != nil {
		log.Printf("ERROR listing npc locations: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list npc locations")
		return
	}
	if links == nil {
		links = []model.NPCLocationLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *RelationshipHandler) LinkNPCLocation(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	var req model.LinkNPCLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.LocationID == 0 {
		writeError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	if err := h.store.LinkNPCLocation(r.Context(), nid, req); err != nil {
		log.Printf("ERROR linking npc %d to location: %v", nid, err)
		writeError(w, http.StatusInternalServerError, "failed to link npc to location")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RelationshipHandler) UnlinkNPCLocation(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	lid, err := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid location id")
		return
	}
	if err := h.store.UnlinkNPCLocation(r.Context(), nid, lid); err != nil {
		log.Printf("ERROR unlinking npc %d from location %d: %v", nid, lid, err)
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Faction <-> Location ---

func (h *RelationshipHandler) ListFactionLocations(w http.ResponseWriter, r *http.Request) {
	fid, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid faction id")
		return
	}
	links, err := h.store.ListFactionLocations(r.Context(), fid)
	if err != nil {
		log.Printf("ERROR listing faction locations: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list faction locations")
		return
	}
	if links == nil {
		links = []model.FactionLocationLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *RelationshipHandler) LinkFactionLocation(w http.ResponseWriter, r *http.Request) {
	fid, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid faction id")
		return
	}
	var req model.LinkFactionLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.LocationID == 0 {
		writeError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	if err := h.store.LinkFactionLocation(r.Context(), fid, req); err != nil {
		log.Printf("ERROR linking faction %d to location: %v", fid, err)
		writeError(w, http.StatusInternalServerError, "failed to link faction to location")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RelationshipHandler) UnlinkFactionLocation(w http.ResponseWriter, r *http.Request) {
	fid, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid faction id")
		return
	}
	lid, err := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid location id")
		return
	}
	if err := h.store.UnlinkFactionLocation(r.Context(), fid, lid); err != nil {
		log.Printf("ERROR unlinking faction %d from location %d: %v", fid, lid, err)
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- NPC <-> NPC ---

func (h *RelationshipHandler) ListNPCRelationships(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	links, err := h.store.ListNPCRelationships(r.Context(), nid)
	if err != nil {
		log.Printf("ERROR listing npc relationships: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list npc relationships")
		return
	}
	if links == nil {
		links = []model.NPCRelationshipLink{}
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *RelationshipHandler) CreateNPCRelationship(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	var req model.CreateNPCRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OtherNPCID == 0 {
		writeError(w, http.StatusBadRequest, "other_npc_id is required")
		return
	}
	if req.Relationship == "" {
		writeError(w, http.StatusBadRequest, "relationship is required")
		return
	}
	if err := h.store.CreateNPCRelationship(r.Context(), nid, req); err != nil {
		log.Printf("ERROR creating npc relationship: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create npc relationship")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RelationshipHandler) DeleteNPCRelationship(w http.ResponseWriter, r *http.Request) {
	nid, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	oid, err := strconv.Atoi(chi.URLParam(r, "otherNPCID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid npc id")
		return
	}
	if err := h.store.DeleteNPCRelationship(r.Context(), nid, oid); err != nil {
		log.Printf("ERROR deleting npc relationship: %v", err)
		writeError(w, http.StatusNotFound, "relationship not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
