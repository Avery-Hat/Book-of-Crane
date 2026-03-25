# Book of Crane

A campaign notes app for tabletop RPGs. Track NPCs, locations, factions, items, and sessions — and the relationships between them.

Built with **Go**, **PostgreSQL**, and **Docker**.

## Quick Start

```bash
git clone https://github.com/Avery-Hat/Book-of-Crane.git
cd Book-of-Crane
make docker-up
```

Open `http://localhost:8080` in your browser. Register an account to start creating and editing data.

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

Set `JWT_SECRET` in your environment for production — a dev default is used if unset.

## Web UI

The web UI is available at `http://localhost:8080`. All pages are publicly readable. Log in to create, edit, and delete data.

| Page | URL |
|------|-----|
| Campaign list | `/campaigns` |
| Campaign detail | `/campaigns/{id}` |
| NPC detail | `/campaigns/{id}/npcs/{id}` |
| Location detail | `/campaigns/{id}/locations/{id}` |
| Faction detail | `/campaigns/{id}/factions/{id}` |
| Session detail | `/campaigns/{id}/sessions/{id}` |
| Search | `/campaigns/{id}/search?q=` |
| Login | `/login` |
| Register | `/register` |

## Authentication

All `GET` endpoints are public. Write operations (`POST`, `PUT`, `DELETE`) require authentication.

**Web UI** — register and log in at `/register` and `/login`. A session cookie is set automatically.

**API** — use the auth endpoints to get a JWT token, then include it on write requests:

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "your-username", "password": "your-password"}'

# Login — returns a token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "your-username", "password": "your-password"}'
# → {"token": "eyJ..."}

# Include token on write requests
curl -X POST http://localhost:8080/api/campaigns \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{"name": "Curse of Strahd"}'
```

Tokens expire after 24 hours. Multiple accounts are supported — all logged-in users share access to all data.

## API

All endpoints are under `/api/`. Write endpoints require an `Authorization: Bearer <token>` header.

### Campaigns

```bash
curl http://localhost:8080/api/campaigns
curl http://localhost:8080/api/campaigns/1

curl -X POST http://localhost:8080/api/campaigns \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "Descent into Avernus", "description": "Elturel has fallen into Avernus"}'

curl -X PUT http://localhost:8080/api/campaigns/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "Descent into Avernus (Complete)"}'

curl -X DELETE http://localhost:8080/api/campaigns/1 \
  -H "Authorization: Bearer <token>"
```

### NPCs

```bash
curl http://localhost:8080/api/campaigns/1/npcs
curl http://localhost:8080/api/campaigns/1/npcs/1
curl http://localhost:8080/api/campaigns/1/npcs/1/detail   # full detail with all connections

curl -X POST http://localhost:8080/api/campaigns/1/npcs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "Reya Mantlemorn", "race": "Human", "role": "Soldier", "status": "alive"}'
```

NPC `status` must be `alive`, `dead`, or `unknown` (case-insensitive).

### Locations

```bash
curl http://localhost:8080/api/campaigns/1/locations

curl -X POST http://localhost:8080/api/campaigns/1/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "Elturel", "type": "City"}'
```

### Factions

```bash
curl http://localhost:8080/api/campaigns/1/factions

curl -X POST http://localhost:8080/api/campaigns/1/factions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "Hellriders", "alignment": "lawful good"}'
```

### Items

```bash
curl http://localhost:8080/api/campaigns/1/items

curl -X POST http://localhost:8080/api/campaigns/1/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name": "Sword of Zariel", "type": "weapon", "rarity": "artifact"}'
```

### Sessions

```bash
curl http://localhost:8080/api/campaigns/1/sessions
curl http://localhost:8080/api/campaigns/1/sessions/1/recap   # full session recap

curl -X POST http://localhost:8080/api/campaigns/1/sessions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"session_number": 1, "title": "A Night in Baldurs Gate", "played_on": "2024-01-15T00:00:00Z"}'
```

### Relationships

```bash
# Link an NPC to a faction
curl -X POST http://localhost:8080/api/campaigns/1/npcs/1/factions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"faction_id": 1, "role": "leader"}'

# Link an NPC to a location
curl -X POST http://localhost:8080/api/campaigns/1/npcs/1/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"location_id": 1, "context": "born here"}'

# Link two NPCs
curl -X POST http://localhost:8080/api/campaigns/1/npcs/1/relationships \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"other_npc_id": 2, "relationship": "siblings"}'

# Link an NPC to a session
curl -X POST http://localhost:8080/api/campaigns/1/sessions/1/npcs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
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
internal/web/        — HTML page handlers and form processing
internal/store/      — database queries
internal/model/      — shared Go structs
internal/middleware/ — request logging, panic recovery, JWT auth
migrations/          — SQL migration files (001–005)
templates/           — Go HTML templates
```

## Tech Stack

- **Go** — chi router, pgx driver, golang-jwt
- **PostgreSQL** — relational data, full-text search via GIN indexes
- **golang-migrate** — migrations run automatically on startup
- **Docker Compose** — one-command local setup
- **Pico CSS** — minimal styling for the web UI
