package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NPCStore struct {
	db *pgxpool.Pool
}

func NewNPCStore(db *pgxpool.Pool) *NPCStore {
	return &NPCStore{db: db}
}

func (s *NPCStore) List(ctx context.Context, campaignID int) ([]model.NPC, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, campaign_id, name, race, role, status, description, notes, created_at, updated_at
		 FROM npcs
		 WHERE campaign_id = $1
		 ORDER BY updated_at DESC`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("querying npcs: %w", err)
	}
	defer rows.Close()

	var npcs []model.NPC
	for rows.Next() {
		var n model.NPC
		if err := rows.Scan(&n.ID, &n.CampaignID, &n.Name, &n.Race, &n.Role, &n.Status, &n.Description, &n.Notes, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning npc: %w", err)
		}
		npcs = append(npcs, n)
	}
	return npcs, rows.Err()
}

func (s *NPCStore) GetByID(ctx context.Context, campaignID, id int) (model.NPC, error) {
	var n model.NPC
	err := s.db.QueryRow(ctx,
		`SELECT id, campaign_id, name, race, role, status, description, notes, created_at, updated_at
		 FROM npcs
		 WHERE id = $1 AND campaign_id = $2`, id, campaignID).
		Scan(&n.ID, &n.CampaignID, &n.Name, &n.Race, &n.Role, &n.Status, &n.Description, &n.Notes, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return n, fmt.Errorf("npc %d not found", id)
		}
		return n, fmt.Errorf("querying npc %d: %w", id, err)
	}
	return n, nil
}

func (s *NPCStore) Create(ctx context.Context, campaignID int, req model.CreateNPCRequest) (model.NPC, error) {
	status := req.Status
	if status == "" {
		status = "alive"
	}
	var n model.NPC
	err := s.db.QueryRow(ctx,
		`INSERT INTO npcs (campaign_id, name, race, role, status, description, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, campaign_id, name, race, role, status, description, notes, created_at, updated_at`,
		campaignID, req.Name, req.Race, req.Role, status, req.Description, req.Notes).
		Scan(&n.ID, &n.CampaignID, &n.Name, &n.Race, &n.Role, &n.Status, &n.Description, &n.Notes, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return n, fmt.Errorf("creating npc: %w", err)
	}
	return n, nil
}

func (s *NPCStore) Update(ctx context.Context, campaignID, id int, req model.UpdateNPCRequest) (model.NPC, error) {
	current, err := s.GetByID(ctx, campaignID, id)
	if err != nil {
		return current, err
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	race := current.Race
	if req.Race != nil {
		race = req.Race
	}
	role := current.Role
	if req.Role != nil {
		role = req.Role
	}
	status := current.Status
	if req.Status != nil {
		status = *req.Status
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}
	notes := current.Notes
	if req.Notes != nil {
		notes = req.Notes
	}

	var n model.NPC
	err = s.db.QueryRow(ctx,
		`UPDATE npcs
		 SET name=$2, race=$3, role=$4, status=$5, description=$6, notes=$7, updated_at=NOW()
		 WHERE id=$1 AND campaign_id=$8
		 RETURNING id, campaign_id, name, race, role, status, description, notes, created_at, updated_at`,
		id, name, race, role, status, description, notes, campaignID).
		Scan(&n.ID, &n.CampaignID, &n.Name, &n.Race, &n.Role, &n.Status, &n.Description, &n.Notes, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return n, fmt.Errorf("updating npc %d: %w", id, err)
	}
	return n, nil
}

func (s *NPCStore) Detail(ctx context.Context, campaignID, id int) (model.NPCDetail, error) {
	var detail model.NPCDetail

	npc, err := s.GetByID(ctx, campaignID, id)
	if err != nil {
		return detail, err
	}
	detail.NPC = npc

	// Factions
	factionRows, err := s.db.Query(ctx,
		`SELECT f.id, f.name, f.alignment, nf.role
		 FROM npc_factions nf
		 JOIN factions f ON f.id = nf.faction_id
		 WHERE nf.npc_id = $1
		 ORDER BY f.name`, id)
	if err != nil {
		return detail, fmt.Errorf("querying npc factions: %w", err)
	}
	defer factionRows.Close()
	var factions []model.NPCFactionLink
	for factionRows.Next() {
		var f model.NPCFactionLink
		if err := factionRows.Scan(&f.FactionID, &f.Name, &f.Alignment, &f.Role); err != nil {
			return detail, fmt.Errorf("scanning npc faction: %w", err)
		}
		factions = append(factions, f)
	}
	if err := factionRows.Err(); err != nil {
		return detail, err
	}
	if factions == nil {
		factions = []model.NPCFactionLink{}
	}
	detail.Factions = factions

	// Locations
	locationRows, err := s.db.Query(ctx,
		`SELECT l.id, l.name, l.type, nl.context
		 FROM npc_locations nl
		 JOIN locations l ON l.id = nl.location_id
		 WHERE nl.npc_id = $1
		 ORDER BY l.name`, id)
	if err != nil {
		return detail, fmt.Errorf("querying npc locations: %w", err)
	}
	defer locationRows.Close()
	var locations []model.NPCLocationLink
	for locationRows.Next() {
		var l model.NPCLocationLink
		if err := locationRows.Scan(&l.LocationID, &l.Name, &l.Type, &l.Context); err != nil {
			return detail, fmt.Errorf("scanning npc location: %w", err)
		}
		locations = append(locations, l)
	}
	if err := locationRows.Err(); err != nil {
		return detail, err
	}
	if locations == nil {
		locations = []model.NPCLocationLink{}
	}
	detail.Locations = locations

	// Relationships
	relRows, err := s.db.Query(ctx,
		`SELECT
		   CASE WHEN nr.npc_id_1 = $1 THEN nr.npc_id_2 ELSE nr.npc_id_1 END,
		   n.name,
		   nr.relationship,
		   nr.notes
		 FROM npc_relationships nr
		 JOIN npcs n ON n.id = CASE WHEN nr.npc_id_1 = $1 THEN nr.npc_id_2 ELSE nr.npc_id_1 END
		 WHERE nr.npc_id_1 = $1 OR nr.npc_id_2 = $1
		 ORDER BY n.name`, id)
	if err != nil {
		return detail, fmt.Errorf("querying npc relationships: %w", err)
	}
	defer relRows.Close()
	var relationships []model.NPCRelationshipLink
	for relRows.Next() {
		var r model.NPCRelationshipLink
		if err := relRows.Scan(&r.NPCID, &r.Name, &r.Relationship, &r.Notes); err != nil {
			return detail, fmt.Errorf("scanning npc relationship: %w", err)
		}
		relationships = append(relationships, r)
	}
	if err := relRows.Err(); err != nil {
		return detail, err
	}
	if relationships == nil {
		relationships = []model.NPCRelationshipLink{}
	}
	detail.Relationships = relationships

	// Sessions
	sessionRows, err := s.db.Query(ctx,
		`SELECT s.id, s.session_number, s.title, sn.introduced
		 FROM session_npcs sn
		 JOIN sessions s ON s.id = sn.session_id
		 WHERE sn.npc_id = $1
		 ORDER BY s.session_number`, id)
	if err != nil {
		return detail, fmt.Errorf("querying npc sessions: %w", err)
	}
	defer sessionRows.Close()
	var sessions []model.NPCSessionLink
	for sessionRows.Next() {
		var sl model.NPCSessionLink
		if err := sessionRows.Scan(&sl.SessionID, &sl.SessionNumber, &sl.Title, &sl.Introduced); err != nil {
			return detail, fmt.Errorf("scanning npc session: %w", err)
		}
		sessions = append(sessions, sl)
	}
	if err := sessionRows.Err(); err != nil {
		return detail, err
	}
	if sessions == nil {
		sessions = []model.NPCSessionLink{}
	}
	detail.Sessions = sessions

	return detail, nil
}

func (s *NPCStore) Delete(ctx context.Context, campaignID, id int) error {
	result, err := s.db.Exec(ctx, `DELETE FROM npcs WHERE id = $1 AND campaign_id = $2`, id, campaignID)
	if err != nil {
		return fmt.Errorf("deleting npc %d: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("npc %d not found", id)
	}
	return nil
}
