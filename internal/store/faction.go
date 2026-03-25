package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FactionStore struct {
	db *pgxpool.Pool
}

func NewFactionStore(db *pgxpool.Pool) *FactionStore {
	return &FactionStore{db: db}
}

func (s *FactionStore) List(ctx context.Context, campaignID int) ([]model.Faction, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, campaign_id, name, alignment, description, notes, created_at, updated_at
		 FROM factions
		 WHERE campaign_id = $1
		 ORDER BY updated_at DESC`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("querying factions: %w", err)
	}
	defer rows.Close()

	var factions []model.Faction
	for rows.Next() {
		var f model.Faction
		if err := rows.Scan(&f.ID, &f.CampaignID, &f.Name, &f.Alignment, &f.Description, &f.Notes, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning faction: %w", err)
		}
		factions = append(factions, f)
	}
	return factions, rows.Err()
}

func (s *FactionStore) GetByID(ctx context.Context, campaignID, id int) (model.Faction, error) {
	var f model.Faction
	err := s.db.QueryRow(ctx,
		`SELECT id, campaign_id, name, alignment, description, notes, created_at, updated_at
		 FROM factions
		 WHERE id = $1 AND campaign_id = $2`, id, campaignID).
		Scan(&f.ID, &f.CampaignID, &f.Name, &f.Alignment, &f.Description, &f.Notes, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return f, fmt.Errorf("faction %d not found", id)
		}
		return f, fmt.Errorf("querying faction %d: %w", id, err)
	}
	return f, nil
}

func (s *FactionStore) Create(ctx context.Context, campaignID int, req model.CreateFactionRequest) (model.Faction, error) {
	var f model.Faction
	err := s.db.QueryRow(ctx,
		`INSERT INTO factions (campaign_id, name, alignment, description, notes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, campaign_id, name, alignment, description, notes, created_at, updated_at`,
		campaignID, req.Name, req.Alignment, req.Description, req.Notes).
		Scan(&f.ID, &f.CampaignID, &f.Name, &f.Alignment, &f.Description, &f.Notes, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return f, fmt.Errorf("creating faction: %w", err)
	}
	return f, nil
}

func (s *FactionStore) Update(ctx context.Context, campaignID, id int, req model.UpdateFactionRequest) (model.Faction, error) {
	current, err := s.GetByID(ctx, campaignID, id)
	if err != nil {
		return current, err
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	alignment := current.Alignment
	if req.Alignment != nil {
		alignment = req.Alignment
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}
	notes := current.Notes
	if req.Notes != nil {
		notes = req.Notes
	}

	var f model.Faction
	err = s.db.QueryRow(ctx,
		`UPDATE factions
		 SET name=$2, alignment=$3, description=$4, notes=$5, updated_at=NOW()
		 WHERE id=$1 AND campaign_id=$6
		 RETURNING id, campaign_id, name, alignment, description, notes, created_at, updated_at`,
		id, name, alignment, description, notes, campaignID).
		Scan(&f.ID, &f.CampaignID, &f.Name, &f.Alignment, &f.Description, &f.Notes, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return f, fmt.Errorf("updating faction %d: %w", id, err)
	}
	return f, nil
}

func (s *FactionStore) Delete(ctx context.Context, campaignID, id int) error {
	result, err := s.db.Exec(ctx, `DELETE FROM factions WHERE id = $1 AND campaign_id = $2`, id, campaignID)
	if err != nil {
		return fmt.Errorf("deleting faction %d: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("faction %d not found", id)
	}
	return nil
}
