package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CampaignStore struct {
	db *pgxpool.Pool
}

func NewCampaignStore(db *pgxpool.Pool) *CampaignStore {
	return &CampaignStore{db: db}
}

func (s *CampaignStore) List(ctx context.Context) ([]model.Campaign, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at
		 FROM campaigns
		 ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []model.Campaign
	for rows.Next() {
		var c model.Campaign
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning campaign: %w", err)
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

func (s *CampaignStore) GetByID(ctx context.Context, id int) (model.Campaign, error) {
	var c model.Campaign
	err := s.db.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at
		 FROM campaigns
		 WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c, fmt.Errorf("campaign %d not found", id)
		}
		return c, fmt.Errorf("querying campaign %d: %w", id, err)
	}
	return c, nil
}

func (s *CampaignStore) Create(ctx context.Context, req model.CreateCampaignRequest) (model.Campaign, error) {
	var c model.Campaign
	err := s.db.QueryRow(ctx,
		`INSERT INTO campaigns (name, description)
		 VALUES ($1, $2)
		 RETURNING id, name, description, created_at, updated_at`,
		req.Name, req.Description).
		Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, fmt.Errorf("creating campaign: %w", err)
	}
	return c, nil
}

func (s *CampaignStore) Update(ctx context.Context, id int, req model.UpdateCampaignRequest) (model.Campaign, error) {
	// Fetch current values so we only update what's provided
	current, err := s.GetByID(ctx, id)
	if err != nil {
		return current, err
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}

	var c model.Campaign
	err = s.db.QueryRow(ctx,
		`UPDATE campaigns
		 SET name = $2, description = $3, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, name, description, created_at, updated_at`,
		id, name, description).
		Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, fmt.Errorf("updating campaign %d: %w", id, err)
	}
	return c, nil
}

func (s *CampaignStore) Delete(ctx context.Context, id int) error {
	result, err := s.db.Exec(ctx, `DELETE FROM campaigns WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting campaign %d: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("campaign %d not found", id)
	}
	return nil
}
