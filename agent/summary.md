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
│   ├── alternatives.go  # GET/POST /api/alternatives, sendJSONError helper
│   └── db.go            # SQLite init, schema, queries
├── frontend/
│   └── index.html       # SPA (inline CSS + JS, no framework)
├── uploads/             # user-uploaded images (gitignored)
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
| `scryfall_id` | TEXT | Scryfall card UUID |
| `filename` | TEXT | Stored filename in uploads/ |
| `uploaded_at` | DATETIME | Default CURRENT_TIMESTAMP |

Index on `scryfall_id`.

## API Endpoints

### GET /api/alternatives?scryfall_id={id}

Returns `{ "alternatives": [{ "id", "url", "uploaded_at" }] }`.

### POST /api/alternatives

Multipart form: `scryfall_id` (string) + `image` (file). Max 5MB, accepts PNG/JPEG/WebP.

Returns 201 with created alternative record.

### GET /uploads/{filename}

Serves uploaded files directly.

### GET /

Serves the embedded SPA from `frontend/`.

## Key Design Decisions

- Frontend is embedded via `go:embed` in `main.go` (not `server/`) because embed paths cannot use `..`
- `Server` receives `fs.FS` for frontend rather than embedding directly in the package
- All error responses are JSON (`{ "error": "..." }`) — frontend always parses JSON
- Scryfall search is called directly from the browser — no server-side proxy needed since Scryfall's API is CORS-friendly and requires no authentication
- Double-faced cards (DFC) detected via `card_faces` — hover flips to show the back
- Upload file cleanup on DB insert failure

## Known Limitations / Future Work

- No pagination on Scryfall results (first page only, up to 175 cards)
- No rate limiting on Scryfall requests (they ask for 50-100ms between requests)
- No authentication or authorization
- No image resizing or optimization
