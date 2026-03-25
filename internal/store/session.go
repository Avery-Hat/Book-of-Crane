package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionStore struct {
	db *pgxpool.Pool
}

func NewSessionStore(db *pgxpool.Pool) *SessionStore {
	return &SessionStore{db: db}
}

// --- CRUD ---

func (s *SessionStore) List(ctx context.Context, campaignID int) ([]model.Session, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, campaign_id, session_number, title, summary, played_on, created_at, updated_at
		 FROM sessions
		 WHERE campaign_id = $1
		 ORDER BY session_number ASC`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		if err := rows.Scan(&s.ID, &s.CampaignID, &s.SessionNumber, &s.Title, &s.Summary, &s.PlayedOn, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (s *SessionStore) GetByID(ctx context.Context, campaignID, id int) (model.Session, error) {
	var sess model.Session
	err := s.db.QueryRow(ctx,
		`SELECT id, campaign_id, session_number, title, summary, played_on, created_at, updated_at
		 FROM sessions
		 WHERE id = $1 AND campaign_id = $2`, id, campaignID).
		Scan(&sess.ID, &sess.CampaignID, &sess.SessionNumber, &sess.Title, &sess.Summary, &sess.PlayedOn, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return sess, fmt.Errorf("session %d not found", id)
		}
		return sess, fmt.Errorf("querying session %d: %w", id, err)
	}
	return sess, nil
}

func (s *SessionStore) Create(ctx context.Context, campaignID int, req model.CreateSessionRequest) (model.Session, error) {
	var sess model.Session
	err := s.db.QueryRow(ctx,
		`INSERT INTO sessions (campaign_id, session_number, title, summary, played_on)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, campaign_id, session_number, title, summary, played_on, created_at, updated_at`,
		campaignID, req.SessionNumber, req.Title, req.Summary, req.PlayedOn).
		Scan(&sess.ID, &sess.CampaignID, &sess.SessionNumber, &sess.Title, &sess.Summary, &sess.PlayedOn, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return sess, fmt.Errorf("creating session: %w", err)
	}
	return sess, nil
}

func (s *SessionStore) Update(ctx context.Context, campaignID, id int, req model.UpdateSessionRequest) (model.Session, error) {
	current, err := s.GetByID(ctx, campaignID, id)
	if err != nil {
		return current, err
	}

	title := current.Title
	if req.Title != nil {
		title = req.Title
	}
	summary := current.Summary
	if req.Summary != nil {
		summary = req.Summary
	}
	playedOn := current.PlayedOn
	if req.PlayedOn != nil {
		playedOn = req.PlayedOn
	}

	var sess model.Session
	err = s.db.QueryRow(ctx,
		`UPDATE sessions
		 SET title=$2, summary=$3, played_on=$4, updated_at=NOW()
		 WHERE id=$1 AND campaign_id=$5
		 RETURNING id, campaign_id, session_number, title, summary, played_on, created_at, updated_at`,
		id, title, summary, playedOn, campaignID).
		Scan(&sess.ID, &sess.CampaignID, &sess.SessionNumber, &sess.Title, &sess.Summary, &sess.PlayedOn, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return sess, fmt.Errorf("updating session %d: %w", id, err)
	}
	return sess, nil
}

func (s *SessionStore) Delete(ctx context.Context, campaignID, id int) error {
	result, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1 AND campaign_id = $2`, id, campaignID)
	if err != nil {
		return fmt.Errorf("deleting session %d: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("session %d not found", id)
	}
	return nil
}

// --- Recap ---

func (s *SessionStore) Recap(ctx context.Context, campaignID, sessionID int) (model.SessionRecap, error) {
	var recap model.SessionRecap

	sess, err := s.GetByID(ctx, campaignID, sessionID)
	if err != nil {
		return recap, err
	}
	recap.Session = sess

	// NPCs with their factions
	npcRows, err := s.db.Query(ctx,
		`SELECT n.id, n.name, n.race, n.role, n.status, sn.introduced
		 FROM session_npcs sn
		 JOIN npcs n ON n.id = sn.npc_id
		 WHERE sn.session_id = $1
		 ORDER BY n.name`, sessionID)
	if err != nil {
		return recap, fmt.Errorf("querying recap npcs: %w", err)
	}
	defer npcRows.Close()

	var recapNPCs []model.RecapNPC
	for npcRows.Next() {
		var n model.RecapNPC
		if err := npcRows.Scan(&n.NPCID, &n.Name, &n.Race, &n.Role, &n.Status, &n.Introduced); err != nil {
			return recap, fmt.Errorf("scanning recap npc: %w", err)
		}
		recapNPCs = append(recapNPCs, n)
	}
	if err := npcRows.Err(); err != nil {
		return recap, err
	}

	// Load factions for each NPC
	for i, n := range recapNPCs {
		factionRows, err := s.db.Query(ctx,
			`SELECT f.id, f.name, f.alignment, nf.role
			 FROM npc_factions nf
			 JOIN factions f ON f.id = nf.faction_id
			 WHERE nf.npc_id = $1
			 ORDER BY f.name`, n.NPCID)
		if err != nil {
			return recap, fmt.Errorf("querying factions for npc %d: %w", n.NPCID, err)
		}
		var factions []model.NPCFactionLink
		for factionRows.Next() {
			var f model.NPCFactionLink
			if err := factionRows.Scan(&f.FactionID, &f.Name, &f.Alignment, &f.Role); err != nil {
				factionRows.Close()
				return recap, fmt.Errorf("scanning faction: %w", err)
			}
			factions = append(factions, f)
		}
		factionRows.Close()
		if err := factionRows.Err(); err != nil {
			return recap, err
		}
		if factions == nil {
			factions = []model.NPCFactionLink{}
		}
		recapNPCs[i].Factions = factions
	}

	if recapNPCs == nil {
		recapNPCs = []model.RecapNPC{}
	}
	recap.NPCs = recapNPCs

	// Locations
	locations, err := s.ListLocations(ctx, sessionID)
	if err != nil {
		return recap, err
	}
	if locations == nil {
		locations = []model.SessionLocationLink{}
	}
	recap.Locations = locations

	// Items
	items, err := s.ListItems(ctx, sessionID)
	if err != nil {
		return recap, err
	}
	if items == nil {
		items = []model.SessionItemLink{}
	}
	recap.Items = items

	return recap, nil
}

// --- Entity links ---

func (s *SessionStore) ListNPCs(ctx context.Context, sessionID int) ([]model.SessionNPCLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT n.id, n.name, n.race, n.role, n.status, sn.introduced
		 FROM session_npcs sn
		 JOIN npcs n ON n.id = sn.npc_id
		 WHERE sn.session_id = $1
		 ORDER BY n.name`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("querying session npcs: %w", err)
	}
	defer rows.Close()

	var links []model.SessionNPCLink
	for rows.Next() {
		var l model.SessionNPCLink
		if err := rows.Scan(&l.NPCID, &l.Name, &l.Race, &l.Role, &l.Status, &l.Introduced); err != nil {
			return nil, fmt.Errorf("scanning session npc: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *SessionStore) LinkNPC(ctx context.Context, sessionID int, req model.LinkSessionNPCRequest) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO session_npcs (session_id, npc_id, introduced)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (session_id, npc_id) DO UPDATE SET introduced = $3`,
		sessionID, req.NPCID, req.Introduced)
	if err != nil {
		return fmt.Errorf("linking npc %d to session %d: %w", req.NPCID, sessionID, err)
	}
	return nil
}

func (s *SessionStore) UnlinkNPC(ctx context.Context, sessionID, npcID int) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM session_npcs WHERE session_id = $1 AND npc_id = $2`, sessionID, npcID)
	if err != nil {
		return fmt.Errorf("unlinking npc %d from session %d: %w", npcID, sessionID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}

func (s *SessionStore) ListLocations(ctx context.Context, sessionID int) ([]model.SessionLocationLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT l.id, l.name, l.type
		 FROM session_locations sl
		 JOIN locations l ON l.id = sl.location_id
		 WHERE sl.session_id = $1
		 ORDER BY l.name`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("querying session locations: %w", err)
	}
	defer rows.Close()

	var links []model.SessionLocationLink
	for rows.Next() {
		var l model.SessionLocationLink
		if err := rows.Scan(&l.LocationID, &l.Name, &l.Type); err != nil {
			return nil, fmt.Errorf("scanning session location: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *SessionStore) LinkLocation(ctx context.Context, sessionID int, req model.LinkSessionLocationRequest) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO session_locations (session_id, location_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		sessionID, req.LocationID)
	if err != nil {
		return fmt.Errorf("linking location %d to session %d: %w", req.LocationID, sessionID, err)
	}
	return nil
}

func (s *SessionStore) UnlinkLocation(ctx context.Context, sessionID, locationID int) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM session_locations WHERE session_id = $1 AND location_id = $2`, sessionID, locationID)
	if err != nil {
		return fmt.Errorf("unlinking location %d from session %d: %w", locationID, sessionID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}

func (s *SessionStore) ListItems(ctx context.Context, sessionID int) ([]model.SessionItemLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT i.id, i.name, i.type, i.rarity
		 FROM session_items si
		 JOIN items i ON i.id = si.item_id
		 WHERE si.session_id = $1
		 ORDER BY i.name`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("querying session items: %w", err)
	}
	defer rows.Close()

	var links []model.SessionItemLink
	for rows.Next() {
		var l model.SessionItemLink
		if err := rows.Scan(&l.ItemID, &l.Name, &l.Type, &l.Rarity); err != nil {
			return nil, fmt.Errorf("scanning session item: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *SessionStore) LinkItem(ctx context.Context, sessionID int, req model.LinkSessionItemRequest) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO session_items (session_id, item_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		sessionID, req.ItemID)
	if err != nil {
		return fmt.Errorf("linking item %d to session %d: %w", req.ItemID, sessionID, err)
	}
	return nil
}

func (s *SessionStore) UnlinkItem(ctx context.Context, sessionID, itemID int) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM session_items WHERE session_id = $1 AND item_id = $2`, sessionID, itemID)
	if err != nil {
		return fmt.Errorf("unlinking item %d from session %d: %w", itemID, sessionID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}
