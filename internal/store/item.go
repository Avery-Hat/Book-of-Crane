package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemStore struct {
	db *pgxpool.Pool
}

func NewItemStore(db *pgxpool.Pool) *ItemStore {
	return &ItemStore{db: db}
}

func (s *ItemStore) List(ctx context.Context, campaignID int) ([]model.Item, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, campaign_id, name, type, rarity, description, notes, created_at, updated_at
		 FROM items
		 WHERE campaign_id = $1
		 ORDER BY updated_at DESC`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("querying items: %w", err)
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var i model.Item
		if err := rows.Scan(&i.ID, &i.CampaignID, &i.Name, &i.Type, &i.Rarity, &i.Description, &i.Notes, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning item: %w", err)
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (s *ItemStore) GetByID(ctx context.Context, campaignID, id int) (model.Item, error) {
	var i model.Item
	err := s.db.QueryRow(ctx,
		`SELECT id, campaign_id, name, type, rarity, description, notes, created_at, updated_at
		 FROM items
		 WHERE id = $1 AND campaign_id = $2`, id, campaignID).
		Scan(&i.ID, &i.CampaignID, &i.Name, &i.Type, &i.Rarity, &i.Description, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return i, fmt.Errorf("item %d not found", id)
		}
		return i, fmt.Errorf("querying item %d: %w", id, err)
	}
	return i, nil
}

func (s *ItemStore) Create(ctx context.Context, campaignID int, req model.CreateItemRequest) (model.Item, error) {
	var i model.Item
	err := s.db.QueryRow(ctx,
		`INSERT INTO items (campaign_id, name, type, rarity, description, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, campaign_id, name, type, rarity, description, notes, created_at, updated_at`,
		campaignID, req.Name, req.Type, req.Rarity, req.Description, req.Notes).
		Scan(&i.ID, &i.CampaignID, &i.Name, &i.Type, &i.Rarity, &i.Description, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return i, fmt.Errorf("creating item: %w", err)
	}
	return i, nil
}

func (s *ItemStore) Update(ctx context.Context, campaignID, id int, req model.UpdateItemRequest) (model.Item, error) {
	current, err := s.GetByID(ctx, campaignID, id)
	if err != nil {
		return current, err
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	itemType := current.Type
	if req.Type != nil {
		itemType = req.Type
	}
	rarity := current.Rarity
	if req.Rarity != nil {
		rarity = req.Rarity
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}
	notes := current.Notes
	if req.Notes != nil {
		notes = req.Notes
	}

	var i model.Item
	err = s.db.QueryRow(ctx,
		`UPDATE items
		 SET name=$2, type=$3, rarity=$4, description=$5, notes=$6, updated_at=NOW()
		 WHERE id=$1 AND campaign_id=$7
		 RETURNING id, campaign_id, name, type, rarity, description, notes, created_at, updated_at`,
		id, name, itemType, rarity, description, notes, campaignID).
		Scan(&i.ID, &i.CampaignID, &i.Name, &i.Type, &i.Rarity, &i.Description, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return i, fmt.Errorf("updating item %d: %w", id, err)
	}
	return i, nil
}

func (s *ItemStore) Delete(ctx context.Context, campaignID, id int) error {
	result, err := s.db.Exec(ctx, `DELETE FROM items WHERE id = $1 AND campaign_id = $2`, id, campaignID)
	if err != nil {
		return fmt.Errorf("deleting item %d: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("item %d not found", id)
	}
	return nil
}
