# Book of Crane

A campaign notes app for tabletop RPGs. Track NPCs, locations, factions, items, and sessions — and the relationships between them.

Built with **Go**, **PostgreSQL**, and **Docker**.

## Quick Start

```bash
git clone https://github.com/Avery-Hat/Book-of-Crane.git
cd Book-of-Crane
make docker-up
```

Open `http://localhost:8080` in your browser.

## Development

```bash
# Start just Postgres
make db-up

# Run the server (migrations run automatically on start)
make run

# Run tests (requires Postgres running)
make test

# Fresh database
make db-reset
```

## Web UI

| Page | URL |
|------|-----|
| Campaign list | `/campaigns` |
| Campaign detail | `/campaigns/{id}` |
| NPC detail | `/campaigns/{id}/npcs/{id}` |
| Location detail | `/campaigns/{id}/locations/{id}` |
| Faction detail | `/campaigns/{id}/factions/{id}` |
| Session detail | `/campaigns/{id}/sessions/{id}` |
| Search | `/campaigns/{id}/search?q=` |

## API

All endpoints are under `/api/`. Data is created and managed via the API; the web UI is read-only.

### Campaigns

```bash
curl -X POST http://localhost:8080/api/campaigns \
  -H "Content-Type: application/json" \
  -d '{"name": "Descent into Avernus", "description": "The holy city of Elturel has disappeared, dragged into the first layer of the Nine Hells"}'

curl http://localhost:8080/api/campaigns
curl http://localhost:8080/api/campaigns/1
curl -X PUT http://localhost:8080/api/campaigns/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Descent into Avernus (In Progress)"}'
curl -X DELETE http://localhost:8080/api/campaigns/1
```

### NPCs

```bash
curl -X POST http://localhost:8080/api/campaigns/1/npcs \
  -H "Content-Type: application/json" \
  -d '{"name": "Reya Mantlemorn", "race": "Human", "role": "Soldier", "status": "Alive"}'

curl http://localhost:8080/api/campaigns/1/npcs
curl http://localhost:8080/api/campaigns/1/npcs/1
curl http://localhost:8080/api/campaigns/1/npcs/1/detail   # full detail with all connections
```

### Locations

```bash
curl -X POST http://localhost:8080/api/campaigns/1/locations \
  -H "Content-Type: application/json" \
  -d '{"name": "Elturel", "type": "City"}'

curl http://localhost:8080/api/campaigns/1/locations
```

### Factions

```bash
curl -X POST http://localhost:8080/api/campaigns/1/factions \
  -H "Content-Type: application/json" \
  -d '{"name": "Hellriders", "alignment": "lawful good"}'

curl http://localhost:8080/api/campaigns/1/factions
```

### Items

```bash
curl -X POST http://localhost:8080/api/campaigns/1/items \
  -H "Content-Type: application/json" \
  -d '{"name": "Sword of Zariel", "type": "weapon", "rarity": "Artifact"}'

curl http://localhost:8080/api/campaigns/1/items
```

### Sessions

```bash
curl -X POST http://localhost:8080/api/campaigns/1/sessions \
  -H "Content-Type: application/json" \
  -d '{"session_number": 1, "title": "Job at Elturel", "played_on": "2023-10-31T00:00:00Z"}'

curl http://localhost:8080/api/campaigns/1/sessions/1/recap  # full session recap
```

### Relationships

```bash
# Link an NPC to a faction
curl -X POST http://localhost:8080/api/campaigns/1/npcs/1/factions \
  -H "Content-Type: application/json" \
  -d '{"faction_id": 1, "role": "leader"}'

# Link an NPC to a location
curl -X POST http://localhost:8080/api/campaigns/1/npcs/1/locations \
  -H "Content-Type: application/json" \
  -d '{"location_id": 1, "context": "born here"}'

# Link two NPCs
curl -X POST http://localhost:8080/api/campaigns/1/npcs/1/relationships \
  -H "Content-Type: application/json" \
  -d '{"other_npc_id": 2, "relationship": "siblings"}'

# Link an NPC to a session
curl -X POST http://localhost:8080/api/campaigns/1/sessions/1/npcs \
  -H "Content-Type: application/json" \
  -d '{"npc_id": 1, "introduced": true}'
```

### Search

```bash
curl "http://localhost:8080/api/campaigns/1/search?q=strahd"
```

Returns matching NPCs, locations, factions, and sessions grouped by type.

## Project Structure

```
cmd/server/          — entry point, router setup
internal/handler/    — HTTP handlers (JSON API)
internal/web/        — HTML page handlers
internal/store/      — database queries
internal/model/      — shared Go structs
internal/middleware/ — request logging
migrations/          — SQL migration files (001–004)
templates/           — Go HTML templates
```

## Tech Stack

- **Go** — chi router, pgx driver
- **PostgreSQL** — relational data, full-text search via GIN indexes
- **golang-migrate** — migrations run automatically on startup
- **Docker Compose** — one-command local setup
- **Pico CSS** — minimal styling for the web UI
