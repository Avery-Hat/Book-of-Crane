package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Avery-Hat/Book-of-Crane/internal/handler"
	"github.com/Avery-Hat/Book-of-Crane/internal/middleware"
	"github.com/Avery-Hat/Book-of-Crane/internal/store"
	"github.com/Avery-Hat/Book-of-Crane/internal/web"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://campaign:campaign@localhost:5433/campaign_notes?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Run migrations
	if err := runMigrations(dbURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations complete")

	// Connect to database
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Set up stores and handlers
	campaignStore := store.NewCampaignStore(pool)
	campaignHandler := handler.NewCampaignHandler(campaignStore)

	npcStore := store.NewNPCStore(pool)
	npcHandler := handler.NewNPCHandler(npcStore)

	locationStore := store.NewLocationStore(pool)
	locationHandler := handler.NewLocationHandler(locationStore)

	factionStore := store.NewFactionStore(pool)
	factionHandler := handler.NewFactionHandler(factionStore)

	itemStore := store.NewItemStore(pool)
	itemHandler := handler.NewItemHandler(itemStore)

	sessionStore := store.NewSessionStore(pool)
	sessionHandler := handler.NewSessionHandler(sessionStore)

	searchStore := store.NewSearchStore(pool)
	searchHandler := handler.NewSearchHandler(searchStore)

	relStore := store.NewRelationshipStore(pool)
	relHandler := handler.NewRelationshipHandler(relStore)

	webHandler := web.NewHandler(
		campaignStore, npcStore, locationStore, factionStore,
		sessionStore, searchStore, relStore, "templates",
	)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Recover)
	r.Use(middleware.Logger)

	// API routes
	r.Mount("/api/campaigns", campaignHandler.Routes())
	r.Mount("/api/campaigns/{campaignID}/npcs", npcHandler.Routes())
	r.Mount("/api/campaigns/{campaignID}/locations", locationHandler.Routes())
	r.Mount("/api/campaigns/{campaignID}/factions", factionHandler.Routes())
	r.Mount("/api/campaigns/{campaignID}/items", itemHandler.Routes())
	r.Mount("/api/campaigns/{campaignID}/sessions", sessionHandler.Routes())
	r.Mount("/api/campaigns/{campaignID}/search", searchHandler.Routes())

	// Relationship routes
	r.Mount("/api/campaigns/{campaignID}/npcs/{npcID}/factions", relHandler.NPCFactionRoutes())
	r.Mount("/api/campaigns/{campaignID}/npcs/{npcID}/locations", relHandler.NPCLocationRoutes())
	r.Mount("/api/campaigns/{campaignID}/npcs/{npcID}/relationships", relHandler.NPCRelationshipRoutes())
	r.Mount("/api/campaigns/{campaignID}/factions/{factionID}/locations", relHandler.FactionLocationRoutes())

	// Web UI
	r.Mount("/", webHandler.Routes())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
