package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocationStore struct {
	db *pgxpool.Pool
}

func NewLocationStore(db *pgxpool.Pool) *LocationStore {
	return &LocationStore{db: db}
}

func (s *LocationStore) List(ctx context.Context, campaignID int) ([]model.Location, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, campaign_id, parent_id, name, type, description, notes, created_at, updated_at
		 FROM locations
		 WHERE campaign_id = $1
		 ORDER BY updated_at DESC`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("querying locations: %w", err)
	}
	defer rows.Close()

	var locations []model.Location
	for rows.Next() {
		var l model.Location
		if err := rows.Scan(&l.ID, &l.CampaignID, &l.ParentID, &l.Name, &l.Type, &l.Description, &l.Notes, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning location: %w", err)
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

func (s *LocationStore) GetByID(ctx context.Context, campaignID, id int) (model.Location, error) {
	var l model.Location
	err := s.db.QueryRow(ctx,
		`SELECT id, campaign_id, parent_id, name, type, description, notes, created_at, updated_at
		 FROM locations
		 WHERE id = $1 AND campaign_id = $2`, id, campaignID).
		Scan(&l.ID, &l.CampaignID, &l.ParentID, &l.Name, &l.Type, &l.Description, &l.Notes, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return l, fmt.Errorf("location %d not found", id)
		}
		return l, fmt.Errorf("querying location %d: %w", id, err)
	}
	return l, nil
}

func (s *LocationStore) Create(ctx context.Context, campaignID int, req model.CreateLocationRequest) (model.Location, error) {
	var l model.Location
	err := s.db.QueryRow(ctx,
		`INSERT INTO locations (campaign_id, parent_id, name, type, description, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, campaign_id, parent_id, name, type, description, notes, created_at, updated_at`,
		campaignID, req.ParentID, req.Name, req.Type, req.Description, req.Notes).
		Scan(&l.ID, &l.CampaignID, &l.ParentID, &l.Name, &l.Type, &l.Description, &l.Notes, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return l, fmt.Errorf("creating location: %w", err)
	}
	return l, nil
}

func (s *LocationStore) Update(ctx context.Context, campaignID, id int, req model.UpdateLocationRequest) (model.Location, error) {
	current, err := s.GetByID(ctx, campaignID, id)
	if err != nil {
		return current, err
	}

	parentID := current.ParentID
	if req.ParentID != nil {
		parentID = req.ParentID
	}
	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	locType := current.Type
	if req.Type != nil {
		locType = req.Type
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}
	notes := current.Notes
	if req.Notes != nil {
		notes = req.Notes
	}

	var l model.Location
	err = s.db.QueryRow(ctx,
		`UPDATE locations
		 SET parent_id=$2, name=$3, type=$4, description=$5, notes=$6, updated_at=NOW()
		 WHERE id=$1 AND campaign_id=$7
		 RETURNING id, campaign_id, parent_id, name, type, description, notes, created_at, updated_at`,
		id, parentID, name, locType, description, notes, campaignID).
		Scan(&l.ID, &l.CampaignID, &l.ParentID, &l.Name, &l.Type, &l.Description, &l.Notes, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return l, fmt.Errorf("updating location %d: %w", id, err)
	}
	return l, nil
}

func (s *LocationStore) Delete(ctx context.Context, campaignID, id int) error {
	result, err := s.db.Exec(ctx, `DELETE FROM locations WHERE id = $1 AND campaign_id = $2`, id, campaignID)
	if err != nil {
		return fmt.Errorf("deleting location %d: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("location %d not found", id)
	}
	return nil
}
