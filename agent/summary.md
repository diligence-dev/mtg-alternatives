# MTG Alternatives — Code Summary

## What It Does

Single Go binary that serves a SPA for searching Magic: The Gathering cards (via Scryfall) and uploading alternative artwork images for any card. Think of it as a community alt-art gallery per card.

## Architecture

```
Browser <---> Go server <---internal---> SQLite (data.db) + uploads/
      |
      +---> Scryfall API (directly from browser, no proxy)
```

- Single Go binary with embedded frontend (via `go:embed`)
- SQLite for metadata, filesystem for uploaded images
- No external database or build tools required
- Browser calls Scryfall API directly (Scryfall sets CORS `Access-Control-Allow-Origin: *`)

## Project Layout

```
mtg-alternatives/
├── main.go              # entry point, embeds frontend/, wires everything
├── server/
│   ├── server.go        # Server struct, route registration, static serving
│   ├── alternatives.go  # GET/POST /api/alternatives, GET /api/cards, sendJSONError helper
│   └── db.go            # SQLite init, schema, migrations, queries
├── frontend/
│   └── index.html       # SPA (inline CSS + JS, no framework)
├── uploads/             # user-uploaded images (gitignored)
├── Dockerfile           # multi-stage build for Fly.io (CGO for sqlite3)
├── fly.toml             # Fly.io deployment config with persistent volume
├── go.mod / go.sum
├── .gitignore
└── agent/               # agent planning docs
```

## Dependencies

- `github.com/mattn/go-sqlite3` — SQLite driver (requires CGO)
- `github.com/google/uuid` — unique filenames for uploads

## Configuration (env vars)

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Listen port |
| `DB_PATH` | `data.db` | SQLite file path |
| `UPLOADS_DIR` | `uploads` | Directory for uploaded images |

## Database Schema

Single `alternatives` table:

| Column | Type | Notes |
|---|---|---|
| `id` | INTEGER PK | Auto-increment |
| `name` | TEXT | Card name (unique identifier) |
| `filename` | TEXT | Stored filename in uploads/ |
| `uploaded_at` | DATETIME | Default CURRENT_TIMESTAMP |

Index on `name`.

Migration: `InitDB` automatically drops the legacy `scryfall_id` column if present (for databases created before the migration to name-based identifiers).

## API Endpoints

### GET /api/alternatives?name={name}

Returns `{ "alternatives": [{ "id", "url", "uploaded_at" }] }`.

### POST /api/alternatives

Multipart form: `name` (string), `image` (file). Max 5MB, accepts PNG/JPEG/WebP.

Returns 201 with created alternative record.

### GET /api/cards?page={page}&limit={limit}

Returns `{ "cards": [{ "name" }], "total": N }` — paginated distinct cards with at least one alternative. Default page=1, limit=30, max limit=100. Ordered by most recent upload first.

### GET /uploads/{filename}

Serves uploaded files directly.

### GET /

Serves the embedded SPA from `frontend/`.

## Deployment (Fly.io)

- `Dockerfile` — multi-stage build: `golang:1.24-alpine` with GCC for CGO, minimal `alpine` runtime
- `fly.toml` — deploys to `ams` region, persistent volume `mtg_data` mounted at `/data`
- Production env: `DB_PATH=/data/data.db`, `UPLOADS_DIR=/data/uploads`
- Auto-stops machines when idle, auto-starts on incoming requests

### Deploy steps

```sh
fly auth login
fly apps create mtg-alternatives
fly volumes create mtg_data --region ams --size 1
fly deploy
```

## Key Design Decisions

- Frontend is embedded via `go:embed` in `main.go` (not `server/`) because embed paths cannot use `..`
- `fs.Sub` is used to strip the `frontend/` prefix so the SPA is served at `/` (not `/frontend/`)
- `Server` receives `fs.FS` for frontend rather than embedding directly in the package
- All error responses are JSON (`{ "error": "..." }`) — frontend always parses JSON
- Scryfall search is called directly from the browser — no server-side proxy needed since Scryfall's API is CORS-friendly and requires no authentication
- Double-faced cards (DFC) detected via `card_faces` — hover flips to show the back
- Upload file cleanup on DB insert failure
- Card names are used as the unique identifier instead of Scryfall IDs. Card images are fetched from Scryfall's named card image URL (`/cards/named?fuzzy={name}&format=image&version=normal`)
- Search results are sent as raw user query to Scryfall (not augmented). Frontend partitions results using `/api/cards` into two sections: cards with alternatives shown first, then a divider "The following cards have no alternatives yet", then remaining cards
- Empty search query shows all cards with alternatives, paginated via `/api/cards` endpoint (page size 30)
- On page load, an empty search is automatically submitted to show the browse view

## Testing

```sh
go test ./server/tests/         # Go backend tests (28 tests)
```

## Known Limitations / Future Work

- No rate limiting on Scryfall requests (they ask for 50-100ms between requests)
- No authentication or authorization
- No image resizing or optimization
