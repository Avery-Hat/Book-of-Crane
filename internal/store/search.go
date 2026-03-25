package store

import (
	"context"
	"fmt"

	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchStore struct {
	db *pgxpool.Pool
}

func NewSearchStore(db *pgxpool.Pool) *SearchStore {
	return &SearchStore{db: db}
}

func (s *SearchStore) Search(ctx context.Context, campaignID int, query string) (model.SearchResults, error) {
	var results model.SearchResults

	npcs, err := s.searchTable(ctx, campaignID, query,
		`SELECT id, name FROM npcs
		 WHERE campaign_id = $1
		   AND to_tsvector('english', coalesce(name,'') || ' ' || coalesce(description,'') || ' ' || coalesce(notes,''))
		       @@ plainto_tsquery('english', $2)
		 ORDER BY name`)
	if err != nil {
		return results, fmt.Errorf("searching npcs: %w", err)
	}
	results.NPCs = npcs

	locations, err := s.searchTable(ctx, campaignID, query,
		`SELECT id, name FROM locations
		 WHERE campaign_id = $1
		   AND to_tsvector('english', coalesce(name,'') || ' ' || coalesce(description,'') || ' ' || coalesce(notes,''))
		       @@ plainto_tsquery('english', $2)
		 ORDER BY name`)
	if err != nil {
		return results, fmt.Errorf("searching locations: %w", err)
	}
	results.Locations = locations

	factions, err := s.searchTable(ctx, campaignID, query,
		`SELECT id, name FROM factions
		 WHERE campaign_id = $1
		   AND to_tsvector('english', coalesce(name,'') || ' ' || coalesce(description,'') || ' ' || coalesce(notes,''))
		       @@ plainto_tsquery('english', $2)
		 ORDER BY name`)
	if err != nil {
		return results, fmt.Errorf("searching factions: %w", err)
	}
	results.Factions = factions

	sessions, err := s.searchTable(ctx, campaignID, query,
		`SELECT id, coalesce(title, 'Session ' || session_number) FROM sessions
		 WHERE campaign_id = $1
		   AND to_tsvector('english', coalesce(title,'') || ' ' || coalesce(summary,''))
		       @@ plainto_tsquery('english', $2)
		 ORDER BY session_number`)
	if err != nil {
		return results, fmt.Errorf("searching sessions: %w", err)
	}
	results.Sessions = sessions

	return results, nil
}

func (s *SearchStore) searchTable(ctx context.Context, campaignID int, query, sql string) ([]model.SearchResult, error) {
	rows, err := s.db.Query(ctx, sql, campaignID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.SearchResult
	for rows.Next() {
		var r model.SearchResult
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if results == nil {
		results = []model.SearchResult{}
	}
	return results, rows.Err()
}
