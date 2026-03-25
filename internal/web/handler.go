package web

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/Avery-Hat/Book-of-Crane/internal/store"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	campaigns *store.CampaignStore
	npcs      *store.NPCStore
	locations *store.LocationStore
	factions  *store.FactionStore
	sessions  *store.SessionStore
	search    *store.SearchStore
	rels      *store.RelationshipStore
	templates string // path to templates directory
}

func NewHandler(
	campaigns *store.CampaignStore,
	npcs *store.NPCStore,
	locations *store.LocationStore,
	factions *store.FactionStore,
	sessions *store.SessionStore,
	search *store.SearchStore,
	rels *store.RelationshipStore,
	templatesDir string,
) *Handler {
	return &Handler{
		campaigns: campaigns,
		npcs:      npcs,
		locations: locations,
		factions:  factions,
		sessions:  sessions,
		search:    search,
		rels:      rels,
		templates: templatesDir,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/campaigns", http.StatusFound)
	})
	r.Get("/campaigns", h.ListCampaigns)
	r.Get("/campaigns/{campaignID}", h.CampaignDetail)
	r.Get("/campaigns/{campaignID}/npcs/{npcID}", h.NPCDetail)
	r.Get("/campaigns/{campaignID}/locations/{locationID}", h.LocationDetail)
	r.Get("/campaigns/{campaignID}/factions/{factionID}", h.FactionDetail)
	r.Get("/campaigns/{campaignID}/sessions/{sessionID}", h.SessionDetail)
	r.Get("/campaigns/{campaignID}/search", h.Search)
	return r
}

func (h *Handler) render(w http.ResponseWriter, page string, data any) {
	funcMap := template.FuncMap{
		"not": func(v any) bool {
			if v == nil {
				return true
			}
			switch val := v.(type) {
			case []model.SearchResult:
				return len(val) == 0
			case bool:
				return !val
			}
			return false
		},
	}

	t, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join(h.templates, "layout.html"),
		filepath.Join(h.templates, page),
	)
	if err != nil {
		log.Printf("ERROR parsing template %s: %v", page, err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ERROR executing template %s: %v", page, err)
	}
}

func cid(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "campaignID"))
}

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.campaigns.List(r.Context())
	if err != nil {
		log.Printf("ERROR listing campaigns: %v", err)
		http.Error(w, "failed to load campaigns", http.StatusInternalServerError)
		return
	}
	h.render(w, "campaigns.html", campaigns)
}

func (h *Handler) CampaignDetail(w http.ResponseWriter, r *http.Request) {
	id, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	campaign, err := h.campaigns.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	npcs, _ := h.npcs.List(ctx, id)
	locations, _ := h.locations.List(ctx, id)
	factions, _ := h.factions.List(ctx, id)
	sessions, _ := h.sessions.List(ctx, id)

	h.render(w, "campaign.html", struct {
		Campaign  model.Campaign
		NPCs      []model.NPC
		Locations []model.Location
		Factions  []model.Faction
		Sessions  []model.Session
	}{campaign, npcs, locations, factions, sessions})
}

func (h *Handler) NPCDetail(w http.ResponseWriter, r *http.Request) {
	campaignID, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	npcID, err := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	detail, err := h.npcs.Detail(r.Context(), campaignID, npcID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, "npc.html", detail)
}

func (h *Handler) LocationDetail(w http.ResponseWriter, r *http.Request) {
	campaignID, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	locationID, err := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	location, err := h.locations.GetByID(ctx, campaignID, locationID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// NPCs linked to this location
	npcs, _ := h.rels.ListLocationNPCs(ctx, locationID)

	h.render(w, "location.html", struct {
		Location model.Location
		NPCs     []model.LocationNPCLink
	}{location, npcs})
}

func (h *Handler) FactionDetail(w http.ResponseWriter, r *http.Request) {
	campaignID, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	factionID, err := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	faction, err := h.factions.GetByID(ctx, campaignID, factionID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	npcs, _ := h.rels.ListFactionNPCs(ctx, factionID)
	locations, _ := h.rels.ListFactionLocations(ctx, factionID)

	h.render(w, "faction.html", struct {
		Faction   model.Faction
		NPCs      []model.FactionNPCLink
		Locations []model.FactionLocationLink
	}{faction, npcs, locations})
}

func (h *Handler) SessionDetail(w http.ResponseWriter, r *http.Request) {
	campaignID, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sessionID, err := strconv.Atoi(chi.URLParam(r, "sessionID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	recap, err := h.sessions.Recap(r.Context(), campaignID, sessionID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.render(w, "session.html", struct {
		Session   model.Session
		NPCs      []model.RecapNPC
		Locations []model.SessionLocationLink
		Items     []model.SessionItemLink
	}{recap.Session, recap.NPCs, recap.Locations, recap.Items})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	campaignID, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))

	data := struct {
		CampaignID int
		Query      string
		Results    *model.SearchResults
	}{CampaignID: campaignID, Query: q}

	if q != "" {
		results, err := h.search.Search(r.Context(), campaignID, q)
		if err != nil {
			log.Printf("ERROR searching: %v", err)
		} else {
			data.Results = &results
		}
	}

	h.render(w, "search.html", data)
}

