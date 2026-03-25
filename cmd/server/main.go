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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
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

	authStore := store.NewAuthStore(pool)
	authHandler := handler.NewAuthHandler(authStore, jwtSecret)

	webHandler := web.NewHandler(
		campaignStore, npcStore, locationStore, factionStore, itemStore,
		sessionStore, searchStore, relStore, authStore, jwtSecret, "templates",
	)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Recover)
	r.Use(middleware.Logger)

	// Auth routes (public)
	r.Mount("/auth", authHandler.Routes())

	// API routes
	requireAuth := middleware.RequireAuth(jwtSecret)
	r.Route("/api", func(api chi.Router) {
		// GET routes — public
		api.Get("/campaigns", campaignHandler.List)
		api.Get("/campaigns/{id}", campaignHandler.Get)
		api.Get("/campaigns/{campaignID}/npcs", npcHandler.List)
		api.Get("/campaigns/{campaignID}/npcs/{npcID}", npcHandler.Get)
		api.Get("/campaigns/{campaignID}/npcs/{npcID}/detail", npcHandler.Detail)
		api.Get("/campaigns/{campaignID}/locations", locationHandler.List)
		api.Get("/campaigns/{campaignID}/locations/{locationID}", locationHandler.Get)
		api.Get("/campaigns/{campaignID}/factions", factionHandler.List)
		api.Get("/campaigns/{campaignID}/factions/{factionID}", factionHandler.Get)
		api.Get("/campaigns/{campaignID}/items", itemHandler.List)
		api.Get("/campaigns/{campaignID}/items/{itemID}", itemHandler.Get)
		api.Get("/campaigns/{campaignID}/sessions", sessionHandler.List)
		api.Get("/campaigns/{campaignID}/sessions/{sessionID}", sessionHandler.Get)
		api.Get("/campaigns/{campaignID}/sessions/{sessionID}/recap", sessionHandler.Recap)
		api.Get("/campaigns/{campaignID}/sessions/{sessionID}/npcs", sessionHandler.ListNPCs)
		api.Get("/campaigns/{campaignID}/sessions/{sessionID}/locations", sessionHandler.ListLocations)
		api.Get("/campaigns/{campaignID}/sessions/{sessionID}/items", sessionHandler.ListItems)
		api.Get("/campaigns/{campaignID}/search", searchHandler.Search)
		api.Get("/campaigns/{campaignID}/npcs/{npcID}/factions", relHandler.ListNPCFactions)
		api.Get("/campaigns/{campaignID}/npcs/{npcID}/locations", relHandler.ListNPCLocations)
		api.Get("/campaigns/{campaignID}/npcs/{npcID}/relationships", relHandler.ListNPCRelationships)
		api.Get("/campaigns/{campaignID}/factions/{factionID}/locations", relHandler.ListFactionLocations)

		// Write routes — require auth
		api.With(requireAuth).Post("/campaigns", campaignHandler.Create)
		api.With(requireAuth).Put("/campaigns/{id}", campaignHandler.Update)
		api.With(requireAuth).Delete("/campaigns/{id}", campaignHandler.Delete)
		api.With(requireAuth).Post("/campaigns/{campaignID}/npcs", npcHandler.Create)
		api.With(requireAuth).Put("/campaigns/{campaignID}/npcs/{npcID}", npcHandler.Update)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/npcs/{npcID}", npcHandler.Delete)
		api.With(requireAuth).Post("/campaigns/{campaignID}/locations", locationHandler.Create)
		api.With(requireAuth).Put("/campaigns/{campaignID}/locations/{locationID}", locationHandler.Update)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/locations/{locationID}", locationHandler.Delete)
		api.With(requireAuth).Post("/campaigns/{campaignID}/factions", factionHandler.Create)
		api.With(requireAuth).Put("/campaigns/{campaignID}/factions/{factionID}", factionHandler.Update)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/factions/{factionID}", factionHandler.Delete)
		api.With(requireAuth).Post("/campaigns/{campaignID}/items", itemHandler.Create)
		api.With(requireAuth).Put("/campaigns/{campaignID}/items/{itemID}", itemHandler.Update)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/items/{itemID}", itemHandler.Delete)
		api.With(requireAuth).Post("/campaigns/{campaignID}/sessions", sessionHandler.Create)
		api.With(requireAuth).Put("/campaigns/{campaignID}/sessions/{sessionID}", sessionHandler.Update)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/sessions/{sessionID}", sessionHandler.Delete)
		api.With(requireAuth).Post("/campaigns/{campaignID}/sessions/{sessionID}/npcs", sessionHandler.LinkNPC)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/sessions/{sessionID}/npcs/{npcID}", sessionHandler.UnlinkNPC)
		api.With(requireAuth).Post("/campaigns/{campaignID}/sessions/{sessionID}/locations", sessionHandler.LinkLocation)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/sessions/{sessionID}/locations/{locationID}", sessionHandler.UnlinkLocation)
		api.With(requireAuth).Post("/campaigns/{campaignID}/sessions/{sessionID}/items", sessionHandler.LinkItem)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/sessions/{sessionID}/items/{itemID}", sessionHandler.UnlinkItem)
		api.With(requireAuth).Post("/campaigns/{campaignID}/npcs/{npcID}/factions", relHandler.LinkNPCFaction)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/npcs/{npcID}/factions/{factionID}", relHandler.UnlinkNPCFaction)
		api.With(requireAuth).Post("/campaigns/{campaignID}/npcs/{npcID}/locations", relHandler.LinkNPCLocation)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/npcs/{npcID}/locations/{locationID}", relHandler.UnlinkNPCLocation)
		api.With(requireAuth).Post("/campaigns/{campaignID}/npcs/{npcID}/relationships", relHandler.CreateNPCRelationship)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/npcs/{npcID}/relationships/{otherNPCID}", relHandler.DeleteNPCRelationship)
		api.With(requireAuth).Post("/campaigns/{campaignID}/factions/{factionID}/locations", relHandler.LinkFactionLocation)
		api.With(requireAuth).Delete("/campaigns/{campaignID}/factions/{factionID}/locations/{locationID}", relHandler.UnlinkFactionLocation)
	})

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
