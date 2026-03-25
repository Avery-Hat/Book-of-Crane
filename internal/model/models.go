package model

import "time"

type Campaign struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCampaignRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateCampaignRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type NPC struct {
	ID          int       `json:"id"`
	CampaignID  int       `json:"campaign_id"`
	Name        string    `json:"name"`
	Race        *string   `json:"race,omitempty"`
	Role        *string   `json:"role,omitempty"`
	Status      string    `json:"status"`
	Description *string   `json:"description,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateNPCRequest struct {
	Name        string  `json:"name"`
	Race        *string `json:"race,omitempty"`
	Role        *string `json:"role,omitempty"`
	Status      string  `json:"status,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type UpdateNPCRequest struct {
	Name        *string `json:"name,omitempty"`
	Race        *string `json:"race,omitempty"`
	Role        *string `json:"role,omitempty"`
	Status      *string `json:"status,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type Location struct {
	ID          int       `json:"id"`
	CampaignID  int       `json:"campaign_id"`
	ParentID    *int      `json:"parent_id,omitempty"`
	Name        string    `json:"name"`
	Type        *string   `json:"type,omitempty"`
	Description *string   `json:"description,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateLocationRequest struct {
	ParentID    *int    `json:"parent_id,omitempty"`
	Name        string  `json:"name"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type UpdateLocationRequest struct {
	ParentID    *int    `json:"parent_id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type Faction struct {
	ID          int       `json:"id"`
	CampaignID  int       `json:"campaign_id"`
	Name        string    `json:"name"`
	Alignment   *string   `json:"alignment,omitempty"`
	Description *string   `json:"description,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateFactionRequest struct {
	Name        string  `json:"name"`
	Alignment   *string `json:"alignment,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type UpdateFactionRequest struct {
	Name        *string `json:"name,omitempty"`
	Alignment   *string `json:"alignment,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type Item struct {
	ID          int       `json:"id"`
	CampaignID  int       `json:"campaign_id"`
	Name        string    `json:"name"`
	Type        *string   `json:"type,omitempty"`
	Rarity      *string   `json:"rarity,omitempty"`
	Description *string   `json:"description,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateItemRequest struct {
	Name        string  `json:"name"`
	Type        *string `json:"type,omitempty"`
	Rarity      *string `json:"rarity,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type UpdateItemRequest struct {
	Name        *string `json:"name,omitempty"`
	Type        *string `json:"type,omitempty"`
	Rarity      *string `json:"rarity,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// --- NPC detail ---

type NPCDetail struct {
	NPC           NPC                   `json:"npc"`
	Factions      []NPCFactionLink      `json:"factions"`
	Locations     []NPCLocationLink     `json:"locations"`
	Relationships []NPCRelationshipLink `json:"relationships"`
	Sessions      []NPCSessionLink      `json:"sessions"`
}

type NPCSessionLink struct {
	SessionID     int     `json:"session_id"`
	SessionNumber int     `json:"session_number"`
	Title         *string `json:"title,omitempty"`
	Introduced    bool    `json:"introduced"`
}

// --- Session recap ---

type SessionRecap struct {
	Session   Session               `json:"session"`
	NPCs      []RecapNPC            `json:"npcs"`
	Locations []SessionLocationLink `json:"locations"`
	Items     []SessionItemLink     `json:"items"`
}

type RecapNPC struct {
	NPCID      int              `json:"npc_id"`
	Name       string           `json:"name"`
	Race       *string          `json:"race,omitempty"`
	Role       *string          `json:"role,omitempty"`
	Status     string           `json:"status"`
	Introduced bool             `json:"introduced"`
	Factions   []NPCFactionLink `json:"factions"`
}

// --- Search ---

type SearchResults struct {
	NPCs      []SearchResult `json:"npcs"`
	Locations []SearchResult `json:"locations"`
	Factions  []SearchResult `json:"factions"`
	Sessions  []SearchResult `json:"sessions"`
}

type SearchResult struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// --- Relationship link types (enriched join results) ---

// Reverse lookups used by location/faction detail pages
type LocationNPCLink struct {
	NPCID   int     `json:"npc_id"`
	Name    string  `json:"name"`
	Context *string `json:"context,omitempty"`
}

type FactionNPCLink struct {
	NPCID int     `json:"npc_id"`
	Name  string  `json:"name"`
	Role  *string `json:"role,omitempty"`
}


type NPCFactionLink struct {
	FactionID  int     `json:"faction_id"`
	Name       string  `json:"name"`
	Alignment  *string `json:"alignment,omitempty"`
	Role       *string `json:"role,omitempty"`
}

type NPCLocationLink struct {
	LocationID int     `json:"location_id"`
	Name       string  `json:"name"`
	Type       *string `json:"type,omitempty"`
	Context    *string `json:"context,omitempty"`
}

type FactionLocationLink struct {
	LocationID   int     `json:"location_id"`
	Name         string  `json:"name"`
	Type         *string `json:"type,omitempty"`
	Relationship *string `json:"relationship,omitempty"`
}

type NPCRelationshipLink struct {
	NPCID        int     `json:"npc_id"`
	Name         string  `json:"name"`
	Relationship string  `json:"relationship"`
	Notes        *string `json:"notes,omitempty"`
}

// --- Session ---

type Session struct {
	ID            int        `json:"id"`
	CampaignID    int        `json:"campaign_id"`
	SessionNumber int        `json:"session_number"`
	Title         *string    `json:"title,omitempty"`
	Summary       *string    `json:"summary,omitempty"`
	PlayedOn      *time.Time `json:"played_on,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateSessionRequest struct {
	SessionNumber int        `json:"session_number"`
	Title         *string    `json:"title,omitempty"`
	Summary       *string    `json:"summary,omitempty"`
	PlayedOn      *time.Time `json:"played_on,omitempty"`
}

type UpdateSessionRequest struct {
	Title    *string    `json:"title,omitempty"`
	Summary  *string    `json:"summary,omitempty"`
	PlayedOn *time.Time `json:"played_on,omitempty"`
}

// Session entity link types (enriched join results)

type SessionNPCLink struct {
	NPCID      int     `json:"npc_id"`
	Name       string  `json:"name"`
	Race       *string `json:"race,omitempty"`
	Role       *string `json:"role,omitempty"`
	Status     string  `json:"status"`
	Introduced bool    `json:"introduced"`
}

type SessionLocationLink struct {
	LocationID int     `json:"location_id"`
	Name       string  `json:"name"`
	Type       *string `json:"type,omitempty"`
}

type SessionItemLink struct {
	ItemID int     `json:"item_id"`
	Name   string  `json:"name"`
	Type   *string `json:"type,omitempty"`
	Rarity *string `json:"rarity,omitempty"`
}

// Session entity link request types

type LinkSessionNPCRequest struct {
	NPCID      int  `json:"npc_id"`
	Introduced bool `json:"introduced"`
}

type LinkSessionLocationRequest struct {
	LocationID int `json:"location_id"`
}

type LinkSessionItemRequest struct {
	ItemID int `json:"item_id"`
}

// --- Auth ---

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// --- Relationship request types ---

type LinkNPCFactionRequest struct {
	FactionID int     `json:"faction_id"`
	Role      *string `json:"role,omitempty"`
}

type LinkNPCLocationRequest struct {
	LocationID int     `json:"location_id"`
	Context    *string `json:"context,omitempty"`
}

type LinkFactionLocationRequest struct {
	LocationID   int     `json:"location_id"`
	Relationship *string `json:"relationship,omitempty"`
}

type CreateNPCRelationshipRequest struct {
	OtherNPCID   int     `json:"other_npc_id"`
	Relationship string  `json:"relationship"`
	Notes        *string `json:"notes,omitempty"`
}
