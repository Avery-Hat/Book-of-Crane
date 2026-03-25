# Book of Crane

A Go REST API for tracking D&D campaign lore — NPCs, locations, factions, and the relationships between them.

Built with **Go**, **PostgreSQL**, and **Docker**.

## Quick Start

```bash
# Clone and start everything
git clone https://github.com/Avery-Hat/Book-of-Crane.git
cd Book-of-Crane
make docker-up
```

The API is available at `http://localhost:8080`.

## Development

```bash
# Start just Postgres
make db-up

# Run the server locally
make run

# Run tests
make test
```

## API Examples

```bash
# Create a campaign
curl -X POST http://localhost:8080/api/campaigns \
  -H "Content-Type: application/json" \
  -d '{"name": "Curse of Strahd", "description": "A gothic horror campaign in Barovia"}'

# List all campaigns
curl http://localhost:8080/api/campaigns

# Get a specific campaign
curl http://localhost:8080/api/campaigns/1

# Update a campaign
curl -X PUT http://localhost:8080/api/campaigns/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Curse of Strahd (Completed)"}'

# Delete a campaign
curl -X DELETE http://localhost:8080/api/campaigns/1
```

## Tech Stack

- **Go** — HTTP server with Chi router
- **PostgreSQL** — Relational data with full-text search
- **golang-migrate** — Database migrations
- **pgx** — PostgreSQL driver
- **Docker Compose** — One-command setup

## Project Structure

```
cmd/server/          → Entry point
internal/handler/    → HTTP handlers
internal/store/      → Database queries (repository pattern)
internal/model/      → Go structs
internal/middleware/  → Request logging
migrations/          → SQL migration files
web/                 → HTML templates and static files
```
