package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RelationshipStore struct {
	db *pgxpool.Pool
}

func NewRelationshipStore(db *pgxpool.Pool) *RelationshipStore {
	return &RelationshipStore{db: db}
}

// --- NPC <-> Faction ---

func (s *RelationshipStore) ListNPCFactions(ctx context.Context, npcID int) ([]model.NPCFactionLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT f.id, f.name, f.alignment, nf.role
		 FROM npc_factions nf
		 JOIN factions f ON f.id = nf.faction_id
		 WHERE nf.npc_id = $1
		 ORDER BY f.name`, npcID)
	if err != nil {
		return nil, fmt.Errorf("querying npc factions: %w", err)
	}
	defer rows.Close()

	var links []model.NPCFactionLink
	for rows.Next() {
		var l model.NPCFactionLink
		if err := rows.Scan(&l.FactionID, &l.Name, &l.Alignment, &l.Role); err != nil {
			return nil, fmt.Errorf("scanning npc faction: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *RelationshipStore) LinkNPCFaction(ctx context.Context, npcID int, req model.LinkNPCFactionRequest) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO npc_factions (npc_id, faction_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (npc_id, faction_id) DO UPDATE SET role = $3`,
		npcID, req.FactionID, req.Role)
	if err != nil {
		return fmt.Errorf("linking npc %d to faction %d: %w", npcID, req.FactionID, err)
	}
	return nil
}

func (s *RelationshipStore) UnlinkNPCFaction(ctx context.Context, npcID, factionID int) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM npc_factions WHERE npc_id = $1 AND faction_id = $2`, npcID, factionID)
	if err != nil {
		return fmt.Errorf("unlinking npc %d from faction %d: %w", npcID, factionID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}

// --- NPC <-> Location ---

func (s *RelationshipStore) ListNPCLocations(ctx context.Context, npcID int) ([]model.NPCLocationLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT l.id, l.name, l.type, nl.context
		 FROM npc_locations nl
		 JOIN locations l ON l.id = nl.location_id
		 WHERE nl.npc_id = $1
		 ORDER BY l.name`, npcID)
	if err != nil {
		return nil, fmt.Errorf("querying npc locations: %w", err)
	}
	defer rows.Close()

	var links []model.NPCLocationLink
	for rows.Next() {
		var l model.NPCLocationLink
		if err := rows.Scan(&l.LocationID, &l.Name, &l.Type, &l.Context); err != nil {
			return nil, fmt.Errorf("scanning npc location: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *RelationshipStore) LinkNPCLocation(ctx context.Context, npcID int, req model.LinkNPCLocationRequest) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO npc_locations (npc_id, location_id, context)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (npc_id, location_id) DO UPDATE SET context = $3`,
		npcID, req.LocationID, req.Context)
	if err != nil {
		return fmt.Errorf("linking npc %d to location %d: %w", npcID, req.LocationID, err)
	}
	return nil
}

func (s *RelationshipStore) UnlinkNPCLocation(ctx context.Context, npcID, locationID int) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM npc_locations WHERE npc_id = $1 AND location_id = $2`, npcID, locationID)
	if err != nil {
		return fmt.Errorf("unlinking npc %d from location %d: %w", npcID, locationID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}

// --- Faction <-> Location ---

func (s *RelationshipStore) ListFactionLocations(ctx context.Context, factionID int) ([]model.FactionLocationLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT l.id, l.name, l.type, fl.relationship
		 FROM faction_locations fl
		 JOIN locations l ON l.id = fl.location_id
		 WHERE fl.faction_id = $1
		 ORDER BY l.name`, factionID)
	if err != nil {
		return nil, fmt.Errorf("querying faction locations: %w", err)
	}
	defer rows.Close()

	var links []model.FactionLocationLink
	for rows.Next() {
		var l model.FactionLocationLink
		if err := rows.Scan(&l.LocationID, &l.Name, &l.Type, &l.Relationship); err != nil {
			return nil, fmt.Errorf("scanning faction location: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *RelationshipStore) LinkFactionLocation(ctx context.Context, factionID int, req model.LinkFactionLocationRequest) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO faction_locations (faction_id, location_id, relationship)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (faction_id, location_id) DO UPDATE SET relationship = $3`,
		factionID, req.LocationID, req.Relationship)
	if err != nil {
		return fmt.Errorf("linking faction %d to location %d: %w", factionID, req.LocationID, err)
	}
	return nil
}

func (s *RelationshipStore) UnlinkFactionLocation(ctx context.Context, factionID, locationID int) error {
	result, err := s.db.Exec(ctx,
		`DELETE FROM faction_locations WHERE faction_id = $1 AND location_id = $2`, factionID, locationID)
	if err != nil {
		return fmt.Errorf("unlinking faction %d from location %d: %w", factionID, locationID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}

// --- Reverse lookups ---

func (s *RelationshipStore) ListLocationNPCs(ctx context.Context, locationID int) ([]model.LocationNPCLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT n.id, n.name, nl.context
		 FROM npc_locations nl
		 JOIN npcs n ON n.id = nl.npc_id
		 WHERE nl.location_id = $1
		 ORDER BY n.name`, locationID)
	if err != nil {
		return nil, fmt.Errorf("querying location npcs: %w", err)
	}
	defer rows.Close()

	var links []model.LocationNPCLink
	for rows.Next() {
		var l model.LocationNPCLink
		if err := rows.Scan(&l.NPCID, &l.Name, &l.Context); err != nil {
			return nil, fmt.Errorf("scanning location npc: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *RelationshipStore) ListFactionNPCs(ctx context.Context, factionID int) ([]model.FactionNPCLink, error) {
	rows, err := s.db.Query(ctx,
		`SELECT n.id, n.name, nf.role
		 FROM npc_factions nf
		 JOIN npcs n ON n.id = nf.npc_id
		 WHERE nf.faction_id = $1
		 ORDER BY n.name`, factionID)
	if err != nil {
		return nil, fmt.Errorf("querying faction npcs: %w", err)
	}
	defer rows.Close()

	var links []model.FactionNPCLink
	for rows.Next() {
		var l model.FactionNPCLink
		if err := rows.Scan(&l.NPCID, &l.Name, &l.Role); err != nil {
			return nil, fmt.Errorf("scanning faction npc: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// --- NPC <-> NPC ---

func (s *RelationshipStore) ListNPCRelationships(ctx context.Context, npcID int) ([]model.NPCRelationshipLink, error) {
	// Query both sides of the relationship (npc_id_1 < npc_id_2 constraint)
	rows, err := s.db.Query(ctx,
		`SELECT
		   CASE WHEN nr.npc_id_1 = $1 THEN nr.npc_id_2 ELSE nr.npc_id_1 END AS other_id,
		   n.name,
		   nr.relationship,
		   nr.notes
		 FROM npc_relationships nr
		 JOIN npcs n ON n.id = CASE WHEN nr.npc_id_1 = $1 THEN nr.npc_id_2 ELSE nr.npc_id_1 END
		 WHERE nr.npc_id_1 = $1 OR nr.npc_id_2 = $1
		 ORDER BY n.name`, npcID)
	if err != nil {
		return nil, fmt.Errorf("querying npc relationships: %w", err)
	}
	defer rows.Close()

	var links []model.NPCRelationshipLink
	for rows.Next() {
		var l model.NPCRelationshipLink
		if err := rows.Scan(&l.NPCID, &l.Name, &l.Relationship, &l.Notes); err != nil {
			return nil, fmt.Errorf("scanning npc relationship: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *RelationshipStore) CreateNPCRelationship(ctx context.Context, npcID int, req model.CreateNPCRelationshipRequest) error {
	// Enforce npc_id_1 < npc_id_2 to satisfy the CHECK constraint
	id1, id2 := npcID, req.OtherNPCID
	if id1 > id2 {
		id1, id2 = id2, id1
	}
	_, err := s.db.Exec(ctx,
		`INSERT INTO npc_relationships (npc_id_1, npc_id_2, relationship, notes)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (npc_id_1, npc_id_2) DO UPDATE SET relationship = $3, notes = $4`,
		id1, id2, req.Relationship, req.Notes)
	if err != nil {
		return fmt.Errorf("creating npc relationship: %w", err)
	}
	return nil
}

func (s *RelationshipStore) DeleteNPCRelationship(ctx context.Context, npcID, otherNPCID int) error {
	id1, id2 := npcID, otherNPCID
	if id1 > id2 {
		id1, id2 = id2, id1
	}
	result, err := s.db.Exec(ctx,
		`DELETE FROM npc_relationships WHERE npc_id_1 = $1 AND npc_id_2 = $2`, id1, id2)
	if err != nil {
		return fmt.Errorf("deleting npc relationship: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("relationship not found")
	}
	return nil
}
