package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/Avery-Hat/Book-of-Crane/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	campaigns *store.CampaignStore
	npcs      *store.NPCStore
	locations *store.LocationStore
	factions  *store.FactionStore
	items     *store.ItemStore
	sessions  *store.SessionStore
	search    *store.SearchStore
	rels      *store.RelationshipStore
	auth      *store.AuthStore
	jwtSecret string
	templates string
}

func NewHandler(
	campaigns *store.CampaignStore,
	npcs *store.NPCStore,
	locations *store.LocationStore,
	factions *store.FactionStore,
	items *store.ItemStore,
	sessions *store.SessionStore,
	search *store.SearchStore,
	rels *store.RelationshipStore,
	auth *store.AuthStore,
	jwtSecret string,
	templatesDir string,
) *Handler {
	return &Handler{
		campaigns: campaigns,
		npcs:      npcs,
		locations: locations,
		factions:  factions,
		items:     items,
		sessions:  sessions,
		search:    search,
		rels:      rels,
		auth:      auth,
		jwtSecret: jwtSecret,
		templates: templatesDir,
	}
}

// --- Page data structs for detail pages with dropdown data ---

type npcPageData struct {
	model.NPCDetail
	AllFactions  []model.Faction
	AllLocations []model.Location
	AllNPCs      []model.NPC
}

type locationPageData struct {
	Location model.Location
	NPCs     []model.LocationNPCLink
	AllNPCs  []model.NPC
}

type factionPageData struct {
	Faction      model.Faction
	NPCs         []model.FactionNPCLink
	Locations    []model.FactionLocationLink
	AllNPCs      []model.NPC
	AllLocations []model.Location
}

type sessionPageData struct {
	Session      model.Session
	NPCs         []model.RecapNPC
	Locations    []model.SessionLocationLink
	Items        []model.SessionItemLink
	AllNPCs      []model.NPC
	AllLocations []model.Location
	AllItems     []model.Item
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
	r.Get("/login", h.GetLogin)
	r.Post("/login", h.PostLogin)
	r.Get("/register", h.GetRegister)
	r.Post("/register", h.PostRegister)
	r.Post("/logout", h.Logout)

	// Write routes (auth checked inside handler)
	r.Post("/campaigns", h.CreateCampaign)
	r.Post("/campaigns/{campaignID}/edit", h.UpdateCampaign)
	r.Post("/campaigns/{campaignID}/delete", h.DeleteCampaign)
	r.Post("/campaigns/{campaignID}/npcs", h.CreateNPC)
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/edit", h.UpdateNPC)
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/delete", h.DeleteNPC)
	r.Post("/campaigns/{campaignID}/locations", h.CreateLocation)
	r.Post("/campaigns/{campaignID}/locations/{locationID}/edit", h.UpdateLocation)
	r.Post("/campaigns/{campaignID}/locations/{locationID}/delete", h.DeleteLocation)
	r.Post("/campaigns/{campaignID}/factions", h.CreateFaction)
	r.Post("/campaigns/{campaignID}/factions/{factionID}/edit", h.UpdateFaction)
	r.Post("/campaigns/{campaignID}/factions/{factionID}/delete", h.DeleteFaction)
	r.Post("/campaigns/{campaignID}/sessions", h.CreateSession)
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/edit", h.UpdateSession)
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/delete", h.DeleteSession)

	// NPC relationship routes
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/link-faction", h.LinkNPCFaction)
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/factions/{factionID}/unlink", h.UnlinkNPCFaction)
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/link-location", h.LinkNPCLocation)
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/locations/{locationID}/unlink", h.UnlinkNPCLocation)
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/link-npc", h.LinkNPCRelationship)
	r.Post("/campaigns/{campaignID}/npcs/{npcID}/relationships/{otherNPCID}/unlink", h.UnlinkNPCRelationship)

	// Location relationship routes
	r.Post("/campaigns/{campaignID}/locations/{locationID}/link-npc", h.LinkLocationNPC)
	r.Post("/campaigns/{campaignID}/locations/{locationID}/npcs/{npcID}/unlink", h.UnlinkLocationNPC)

	// Faction relationship routes
	r.Post("/campaigns/{campaignID}/factions/{factionID}/link-npc", h.LinkFactionNPC)
	r.Post("/campaigns/{campaignID}/factions/{factionID}/npcs/{npcID}/unlink", h.UnlinkFactionNPC)
	r.Post("/campaigns/{campaignID}/factions/{factionID}/link-location", h.LinkFactionLocation)
	r.Post("/campaigns/{campaignID}/factions/{factionID}/locations/{locationID}/unlink", h.UnlinkFactionLocation)

	// Session relationship routes
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/link-npc", h.LinkSessionNPC)
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/npcs/{npcID}/unlink", h.UnlinkSessionNPC)
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/link-location", h.LinkSessionLocation)
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/locations/{locationID}/unlink", h.UnlinkSessionLocation)
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/link-item", h.LinkSessionItem)
	r.Post("/campaigns/{campaignID}/sessions/{sessionID}/items/{itemID}/unlink", h.UnlinkSessionItem)

	return r
}

// checkCookie reads the JWT cookie and returns (loggedIn, username).
func (h *Handler) checkCookie(r *http.Request) (bool, string) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return false, ""
	}
	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return false, ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, ""
	}
	username, _ := claims["username"].(string)
	return true, username
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, page string, data any) {
	loggedIn, username := h.checkCookie(r)

	funcMap := template.FuncMap{
		"loggedIn":    func() bool { return loggedIn },
		"currentUser": func() string { return username },
		"lower":       strings.ToLower,
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

// --- Page handlers ---

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.campaigns.List(r.Context())
	if err != nil {
		log.Printf("ERROR listing campaigns: %v", err)
		http.Error(w, "failed to load campaigns", http.StatusInternalServerError)
		return
	}
	h.render(w, r, "campaigns.html", campaigns)
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

	h.render(w, r, "campaign.html", struct {
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
	ctx := r.Context()
	detail, err := h.npcs.Detail(ctx, campaignID, npcID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	allFactions, _ := h.factions.List(ctx, campaignID)
	allLocations, _ := h.locations.List(ctx, campaignID)
	allNPCs, _ := h.npcs.List(ctx, campaignID)
	h.render(w, r, "npc.html", npcPageData{
		NPCDetail:    detail,
		AllFactions:  allFactions,
		AllLocations: allLocations,
		AllNPCs:      allNPCs,
	})
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

	npcs, _ := h.rels.ListLocationNPCs(ctx, locationID)
	allNPCs, _ := h.npcs.List(ctx, campaignID)

	h.render(w, r, "location.html", locationPageData{
		Location: location,
		NPCs:     npcs,
		AllNPCs:  allNPCs,
	})
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
	allNPCs, _ := h.npcs.List(ctx, campaignID)
	allLocations, _ := h.locations.List(ctx, campaignID)

	h.render(w, r, "faction.html", factionPageData{
		Faction:      faction,
		NPCs:         npcs,
		Locations:    locations,
		AllNPCs:      allNPCs,
		AllLocations: allLocations,
	})
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

	ctx := r.Context()
	recap, err := h.sessions.Recap(ctx, campaignID, sessionID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	allNPCs, _ := h.npcs.List(ctx, campaignID)
	allLocations, _ := h.locations.List(ctx, campaignID)
	allItems, _ := h.items.List(ctx, campaignID)

	h.render(w, r, "session.html", sessionPageData{
		Session:      recap.Session,
		NPCs:         recap.NPCs,
		Locations:    recap.Locations,
		Items:        recap.Items,
		AllNPCs:      allNPCs,
		AllLocations: allLocations,
		AllItems:     allItems,
	})
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

	h.render(w, r, "search.html", data)
}

// --- Auth page handlers ---

type authPageData struct {
	Error string
}

func (h *Handler) GetLogin(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "login.html", authPageData{})
}

func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	hash, err := h.auth.GetPasswordHash(r.Context(), username)
	if err != nil {
		h.render(w, r, "login.html", authPageData{Error: "Invalid username or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		h.render(w, r, "login.html", authPageData{Error: "Invalid username or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		log.Printf("ERROR signing token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})
	http.Redirect(w, r, "/campaigns", http.StatusSeeOther)
}

func (h *Handler) GetRegister(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "register.html", authPageData{})
}

func (h *Handler) PostRegister(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" {
		h.render(w, r, "register.html", authPageData{Error: "Username is required"})
		return
	}
	if len(password) < 8 {
		h.render(w, r, "register.html", authPageData{Error: "Password must be at least 8 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR hashing password: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	if err := h.auth.CreateUser(r.Context(), username, string(hash)); err != nil {
		if err == store.ErrUsernameTaken {
			h.render(w, r, "register.html", authPageData{Error: "Username already taken"})
			return
		}
		log.Printf("ERROR creating user: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/campaigns", http.StatusSeeOther)
}

// --- Write helpers ---

// requireWebAuth redirects to /login if the user is not logged in.
// Returns false if the request should stop.
func (h *Handler) requireWebAuth(w http.ResponseWriter, r *http.Request) bool {
	loggedIn, _ := h.checkCookie(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	return true
}

// strPtr returns nil for empty strings, otherwise a pointer to the value.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// --- Campaign writes ---

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Redirect(w, r, "/campaigns", http.StatusSeeOther)
		return
	}
	req := model.CreateCampaignRequest{Name: name, Description: strPtr(r.FormValue("description"))}
	c, err := h.campaigns.Create(r.Context(), req)
	if err != nil {
		log.Printf("ERROR creating campaign: %v", err)
		http.Error(w, "failed to create campaign", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", c.ID), http.StatusSeeOther)
}

func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	id, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := r.FormValue("name")
	req := model.UpdateCampaignRequest{Name: strPtr(name), Description: strPtr(r.FormValue("description"))}
	if _, err := h.campaigns.Update(r.Context(), id, req); err != nil {
		log.Printf("ERROR updating campaign %d: %v", id, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
}

func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	id, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.campaigns.Delete(r.Context(), id); err != nil {
		log.Printf("ERROR deleting campaign %d: %v", id, err)
	}
	http.Redirect(w, r, "/campaigns", http.StatusSeeOther)
}

// --- NPC writes ---

func (h *Handler) CreateNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	id, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
		return
	}
	status := r.FormValue("status")
	if status == "" {
		status = "alive"
	}
	req := model.CreateNPCRequest{
		Name:        name,
		Race:        strPtr(r.FormValue("race")),
		Role:        strPtr(r.FormValue("role")),
		Status:      status,
		Description: strPtr(r.FormValue("description")),
	}
	if _, err := h.npcs.Create(r.Context(), id, req); err != nil {
		log.Printf("ERROR creating npc: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
}

func (h *Handler) UpdateNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	status := r.FormValue("status")
	req := model.UpdateNPCRequest{
		Name:        strPtr(r.FormValue("name")),
		Race:        strPtr(r.FormValue("race")),
		Role:        strPtr(r.FormValue("role")),
		Status:      strPtr(status),
		Description: strPtr(r.FormValue("description")),
		Notes:       strPtr(r.FormValue("notes")),
	}
	if _, err := h.npcs.Update(r.Context(), campaignID, npcID, req); err != nil {
		log.Printf("ERROR updating npc %d: %v", npcID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/npcs/%d", campaignID, npcID), http.StatusSeeOther)
}

func (h *Handler) DeleteNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	if err := h.npcs.Delete(r.Context(), campaignID, npcID); err != nil {
		log.Printf("ERROR deleting npc %d: %v", npcID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", campaignID), http.StatusSeeOther)
}

// --- Location writes ---

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	id, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
		return
	}
	req := model.CreateLocationRequest{
		Name:        name,
		Type:        strPtr(r.FormValue("type")),
		Description: strPtr(r.FormValue("description")),
	}
	if _, err := h.locations.Create(r.Context(), id, req); err != nil {
		log.Printf("ERROR creating location: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	req := model.UpdateLocationRequest{
		Name:        strPtr(r.FormValue("name")),
		Type:        strPtr(r.FormValue("type")),
		Description: strPtr(r.FormValue("description")),
		Notes:       strPtr(r.FormValue("notes")),
	}
	if _, err := h.locations.Update(r.Context(), campaignID, locationID, req); err != nil {
		log.Printf("ERROR updating location %d: %v", locationID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/locations/%d", campaignID, locationID), http.StatusSeeOther)
}

func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	if err := h.locations.Delete(r.Context(), campaignID, locationID); err != nil {
		log.Printf("ERROR deleting location %d: %v", locationID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", campaignID), http.StatusSeeOther)
}

// --- Faction writes ---

func (h *Handler) CreateFaction(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	id, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
		return
	}
	req := model.CreateFactionRequest{
		Name:        name,
		Alignment:   strPtr(r.FormValue("alignment")),
		Description: strPtr(r.FormValue("description")),
	}
	if _, err := h.factions.Create(r.Context(), id, req); err != nil {
		log.Printf("ERROR creating faction: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
}

func (h *Handler) UpdateFaction(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	req := model.UpdateFactionRequest{
		Name:        strPtr(r.FormValue("name")),
		Alignment:   strPtr(r.FormValue("alignment")),
		Description: strPtr(r.FormValue("description")),
		Notes:       strPtr(r.FormValue("notes")),
	}
	if _, err := h.factions.Update(r.Context(), campaignID, factionID, req); err != nil {
		log.Printf("ERROR updating faction %d: %v", factionID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/factions/%d", campaignID, factionID), http.StatusSeeOther)
}

func (h *Handler) DeleteFaction(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	if err := h.factions.Delete(r.Context(), campaignID, factionID); err != nil {
		log.Printf("ERROR deleting faction %d: %v", factionID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", campaignID), http.StatusSeeOther)
}

// --- Session writes ---

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	id, err := cid(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sessionNum, err := strconv.Atoi(r.FormValue("session_number"))
	if err != nil || sessionNum < 1 {
		http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
		return
	}
	req := model.CreateSessionRequest{
		SessionNumber: sessionNum,
		Title:         strPtr(r.FormValue("title")),
		Summary:       strPtr(r.FormValue("summary")),
	}
	if dateStr := r.FormValue("played_on"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			req.PlayedOn = &t
		}
	}
	if _, err := h.sessions.Create(r.Context(), id, req); err != nil {
		log.Printf("ERROR creating session: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", id), http.StatusSeeOther)
}

func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	req := model.UpdateSessionRequest{
		Title:   strPtr(r.FormValue("title")),
		Summary: strPtr(r.FormValue("summary")),
	}
	if dateStr := r.FormValue("played_on"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			req.PlayedOn = &t
		}
	}
	if _, err := h.sessions.Update(r.Context(), campaignID, sessionID, req); err != nil {
		log.Printf("ERROR updating session %d: %v", sessionID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/sessions/%d", campaignID, sessionID), http.StatusSeeOther)
}

// --- NPC relationship handlers ---

func (h *Handler) LinkNPCFaction(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	factionID, _ := strconv.Atoi(r.FormValue("faction_id"))
	if factionID != 0 {
		req := model.LinkNPCFactionRequest{FactionID: factionID, Role: strPtr(r.FormValue("role"))}
		if err := h.rels.LinkNPCFaction(r.Context(), npcID, req); err != nil {
			log.Printf("ERROR linking npc faction: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/npcs/%d", campaignID, npcID), http.StatusSeeOther)
}

func (h *Handler) UnlinkNPCFaction(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	factionID, _ := strconv.Atoi(chi.URLParam(r, "factionID"))
	if err := h.rels.UnlinkNPCFaction(r.Context(), npcID, factionID); err != nil {
		log.Printf("ERROR unlinking npc faction: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/npcs/%d", campaignID, npcID), http.StatusSeeOther)
}

func (h *Handler) LinkNPCLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	locationID, _ := strconv.Atoi(r.FormValue("location_id"))
	if locationID != 0 {
		req := model.LinkNPCLocationRequest{LocationID: locationID, Context: strPtr(r.FormValue("context"))}
		if err := h.rels.LinkNPCLocation(r.Context(), npcID, req); err != nil {
			log.Printf("ERROR linking npc location: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/npcs/%d", campaignID, npcID), http.StatusSeeOther)
}

func (h *Handler) UnlinkNPCLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	locationID, _ := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err := h.rels.UnlinkNPCLocation(r.Context(), npcID, locationID); err != nil {
		log.Printf("ERROR unlinking npc location: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/npcs/%d", campaignID, npcID), http.StatusSeeOther)
}

func (h *Handler) LinkNPCRelationship(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	otherID, _ := strconv.Atoi(r.FormValue("other_npc_id"))
	relationship := r.FormValue("relationship")
	if otherID != 0 && relationship != "" {
		req := model.CreateNPCRelationshipRequest{OtherNPCID: otherID, Relationship: relationship, Notes: strPtr(r.FormValue("notes"))}
		if err := h.rels.CreateNPCRelationship(r.Context(), npcID, req); err != nil {
			log.Printf("ERROR linking npc relationship: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/npcs/%d", campaignID, npcID), http.StatusSeeOther)
}

func (h *Handler) UnlinkNPCRelationship(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	otherID, _ := strconv.Atoi(chi.URLParam(r, "otherNPCID"))
	if err := h.rels.DeleteNPCRelationship(r.Context(), npcID, otherID); err != nil {
		log.Printf("ERROR unlinking npc relationship: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/npcs/%d", campaignID, npcID), http.StatusSeeOther)
}

// --- Location relationship handlers ---

func (h *Handler) LinkLocationNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	locationID, _ := strconv.Atoi(chi.URLParam(r, "locationID"))
	npcID, _ := strconv.Atoi(r.FormValue("npc_id"))
	if npcID != 0 {
		req := model.LinkNPCLocationRequest{LocationID: locationID, Context: strPtr(r.FormValue("context"))}
		if err := h.rels.LinkNPCLocation(r.Context(), npcID, req); err != nil {
			log.Printf("ERROR linking location npc: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/locations/%d", campaignID, locationID), http.StatusSeeOther)
}

func (h *Handler) UnlinkLocationNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	locationID, _ := strconv.Atoi(chi.URLParam(r, "locationID"))
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err := h.rels.UnlinkNPCLocation(r.Context(), npcID, locationID); err != nil {
		log.Printf("ERROR unlinking location npc: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/locations/%d", campaignID, locationID), http.StatusSeeOther)
}

// --- Faction relationship handlers ---

func (h *Handler) LinkFactionNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	factionID, _ := strconv.Atoi(chi.URLParam(r, "factionID"))
	npcID, _ := strconv.Atoi(r.FormValue("npc_id"))
	if npcID != 0 {
		req := model.LinkNPCFactionRequest{FactionID: factionID, Role: strPtr(r.FormValue("role"))}
		if err := h.rels.LinkNPCFaction(r.Context(), npcID, req); err != nil {
			log.Printf("ERROR linking faction npc: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/factions/%d", campaignID, factionID), http.StatusSeeOther)
}

func (h *Handler) UnlinkFactionNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	factionID, _ := strconv.Atoi(chi.URLParam(r, "factionID"))
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err := h.rels.UnlinkNPCFaction(r.Context(), npcID, factionID); err != nil {
		log.Printf("ERROR unlinking faction npc: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/factions/%d", campaignID, factionID), http.StatusSeeOther)
}

func (h *Handler) LinkFactionLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	factionID, _ := strconv.Atoi(chi.URLParam(r, "factionID"))
	locationID, _ := strconv.Atoi(r.FormValue("location_id"))
	if locationID != 0 {
		req := model.LinkFactionLocationRequest{LocationID: locationID, Relationship: strPtr(r.FormValue("relationship"))}
		if err := h.rels.LinkFactionLocation(r.Context(), factionID, req); err != nil {
			log.Printf("ERROR linking faction location: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/factions/%d", campaignID, factionID), http.StatusSeeOther)
}

func (h *Handler) UnlinkFactionLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	factionID, _ := strconv.Atoi(chi.URLParam(r, "factionID"))
	locationID, _ := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err := h.rels.UnlinkFactionLocation(r.Context(), factionID, locationID); err != nil {
		log.Printf("ERROR unlinking faction location: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/factions/%d", campaignID, factionID), http.StatusSeeOther)
}

// --- Session relationship handlers ---

func (h *Handler) LinkSessionNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionID"))
	npcID, _ := strconv.Atoi(r.FormValue("npc_id"))
	if npcID != 0 {
		introduced := r.FormValue("introduced") == "on"
		req := model.LinkSessionNPCRequest{NPCID: npcID, Introduced: introduced}
		if err := h.sessions.LinkNPC(r.Context(), sessionID, req); err != nil {
			log.Printf("ERROR linking session npc: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/sessions/%d", campaignID, sessionID), http.StatusSeeOther)
}

func (h *Handler) UnlinkSessionNPC(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionID"))
	npcID, _ := strconv.Atoi(chi.URLParam(r, "npcID"))
	if err := h.sessions.UnlinkNPC(r.Context(), sessionID, npcID); err != nil {
		log.Printf("ERROR unlinking session npc: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/sessions/%d", campaignID, sessionID), http.StatusSeeOther)
}

func (h *Handler) LinkSessionLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionID"))
	locationID, _ := strconv.Atoi(r.FormValue("location_id"))
	if locationID != 0 {
		req := model.LinkSessionLocationRequest{LocationID: locationID}
		if err := h.sessions.LinkLocation(r.Context(), sessionID, req); err != nil {
			log.Printf("ERROR linking session location: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/sessions/%d", campaignID, sessionID), http.StatusSeeOther)
}

func (h *Handler) UnlinkSessionLocation(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionID"))
	locationID, _ := strconv.Atoi(chi.URLParam(r, "locationID"))
	if err := h.sessions.UnlinkLocation(r.Context(), sessionID, locationID); err != nil {
		log.Printf("ERROR unlinking session location: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/sessions/%d", campaignID, sessionID), http.StatusSeeOther)
}

func (h *Handler) LinkSessionItem(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionID"))
	itemID, _ := strconv.Atoi(r.FormValue("item_id"))
	if itemID != 0 {
		req := model.LinkSessionItemRequest{ItemID: itemID}
		if err := h.sessions.LinkItem(r.Context(), sessionID, req); err != nil {
			log.Printf("ERROR linking session item: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/sessions/%d", campaignID, sessionID), http.StatusSeeOther)
}

func (h *Handler) UnlinkSessionItem(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
	campaignID, _ := cid(r)
	sessionID, _ := strconv.Atoi(chi.URLParam(r, "sessionID"))
	itemID, _ := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err := h.sessions.UnlinkItem(r.Context(), sessionID, itemID); err != nil {
		log.Printf("ERROR unlinking session item: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d/sessions/%d", campaignID, sessionID), http.StatusSeeOther)
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if !h.requireWebAuth(w, r) {
		return
	}
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
	if err := h.sessions.Delete(r.Context(), campaignID, sessionID); err != nil {
		log.Printf("ERROR deleting session %d: %v", sessionID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%d", campaignID), http.StatusSeeOther)
}
