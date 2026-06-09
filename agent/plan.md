# MTG Alternatives — Implementation Plan

## Architecture Overview

Single Go binary serving an embedded SPA (HTML/CSS/JS) and a JSON API. Uploaded images and metadata are stored on disk using SQLite — no external database dependency. Scryfall's public API is used for card search; results are proxied through the backend to avoid CORS issues.

```
Browser <---> Go server <---internal---> SQLite + /uploads/
                      <---external---> Scryfall API
```

---

## Step 1 — Project Skeleton

Create the Go module and directory layout:

```
mtg-alternatives/
├── main.go              # entry point, wires everything together
├── server/
│   ├── server.go        # HTTP server setup, routes
│   ├── search.go        # /api/search handler (Scryfall proxy)
│   ├── alternatives.go  # GET/POST /api/alternatives handlers
│   └── db.go            # SQLite init, queries
├── frontend/
│   └── index.html       # single-page app (HTML + inline CSS + inline JS)
├── uploads/             # user-uploaded alternative images (gitignored)
└── go.mod
```

- `go mod init github.com/diligence-dev/mtg-alternatives`
- Add dependencies: `github.com/mattn/go-sqlite3`
- Create `uploads/` directory, add `uploads/` to `.gitignore`

## Step 2 — Database Layer (`server/db.go`)

Open/create a SQLite database file (e.g. `data.db`) on startup.

**Schema — single table:**

```sql
CREATE TABLE IF NOT EXISTS alternatives (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    scryfall_id TEXT NOT NULL,       -- Scryfall card UUID
    filename   TEXT NOT NULL,        -- stored filename in uploads/
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_scryfall_id ON alternatives(scryfall_id);
```

**Functions to expose:**

| Function | Purpose |
|---|---|
| `InitDB(path string) *sql.DB` | Open DB, run migrations, return handle |
| `GetAlternatives(db, scryfallID) []Alternative` | Return all alternatives for a card |
| `InsertAlternative(db, scryfallID, filename) error` | Save a new alternative record |

`Alternative` struct: `{ ID int, ScryfallID string, Filename string, UploadedAt string }`

## Step 3 — Scryfall Search Proxy (`server/search.go`)

**Endpoint:** `GET /api/search?q={query}`

1. Read `q` query param, validate it's non-empty (return 400 if empty).
2. Send `GET https://api.scryfall.com/cards/search?q={query}` with a short timeout (5s).
3. Read the JSON response. Extract only the fields the frontend needs into a simplified struct:
   - `id` (string) — Scryfall UUID, used as key for alternatives
   - `name` (string)
   - `image_uris` → `normal` (string) — card image URL
4. Return the simplified card array as JSON: `{ "cards": [...] }`.
5. If Scryfall returns an error (no results, bad query), forward a user-friendly error message with appropriate HTTP status.

Scryfall returns a paginated list; for MVP, return only the first page (up to ~175 cards).

## Step 4 — Alternatives API (`server/alternatives.go`)

### GET /api/alternatives?scryfall_id={id}

1. Validate `scryfall_id` is non-empty.
2. Call `GetAlternatives(db, scryfallID)`.
3. Return JSON: `{ "alternatives": [{ "id", "url": "/uploads/{filename}", "uploaded_at" }, ...] }`.

### POST /api/alternatives

 multipart/form-data with fields:
- `scryfall_id` (string, required)
- `image` (file, required — accept `image/png`, `image/jpeg`, `image/webp`)

Steps:
1. Parse multipart form (max size 5 MB).
2. Validate `scryfall_id` is present, validate file type and size.
3. Generate a unique filename: `{uuid}.{ext}` — prevents collisions.
4. Save file to `uploads/` directory.
5. Call `InsertAlternative(db, scryfallID, filename)`.
6. Return 201 with the created alternative record as JSON.

## Step 5 — Static File Serving & Routing (`server/server.go`)

**Routes:**

| Method | Path | Handler |
|---|---|---|
| GET | `/api/search` | Search proxy |
| GET | `/api/alternatives` | List alternatives |
| POST | `/api/alternatives` | Upload alternative |
| GET | `/uploads/` | Serve uploaded files (`http.FileServer`) |
| GET | `/` | Serve `frontend/index.html` (embedded via `//go:embed`) |

Use Go's `embed` package to embed `frontend/index.html` into the binary so deployment is a single file.

**Server struct:**

```go
type Server struct {
    db *sql.DB
    mux *http.ServeMux
}
```

`NewServer(db, uploadsDir) *Server` — registers all routes and returns the server.

## Step 6 — Entry Point (`main.go`)

1. Read config from environment variables (with sensible defaults):

   | Variable | Default | Purpose |
   |---|---|---|
   | `PORT` | `8080` | Listen port |
   | `DB_PATH` | `data.db` | SQLite file path |
   | `UPLOADS_DIR` | `uploads` | Directory for uploaded images |

2. Create `uploads/` dir if it doesn't exist (`os.MkdirAll`).
3. Call `db.InitDB(DB_PATH)`.
4. Create `Server` and start listening on `:PORT`.
5. Log startup message with port.

## Step 7 — Frontend (`frontend/index.html`)

Single HTML file with inline `<style>` and `<script>`. No build tools, no frameworks — plain vanilla JS. Keep it minimal and readable.

### Layout (three states/views in one page)

```
┌─────────────────────────────────────────┐
│  [ Search input ]          [ Search ]   │
├─────────────────────────────────────────┤
│  Card Results Gallery                   │
│  [img] [img] [img] [img] [img]          │
│  [img] [img] [img] [img] [img]          │
│  [img] [img] [img] [img] [img]          │
│  ...                                    │
│  (click to select)                      │
├─────────────────────────────────────────┤
│  Selected Card: <name>                  │
│  Alternatives Gallery                   │
│  [img] [img] [img] ...                  │
│  [ Upload new alternative ]             │
└─────────────────────────────────────────┘
```

### HTML Structure

- `<input id="search-input">` + `<button id="search-btn">`
- `<div id="results">` — card grid (populated by JS)
- `<div id="detail">` — hidden until a card is selected
  - `<h2 id="card-name">`
  - `<div id="alternatives">` — alternative images grid
  - `<form id="upload-form">` — file input + submit button

### JavaScript Logic

1. **Search:** On button click or Enter key, `GET /api/search?q=...`. Render response cards as a clickable image grid. Each card stores its `scryfall_id` and `name` in `data-*` attributes.

2. **Select card:** On card click, set selected card, show detail section. Fetch alternatives: `GET /api/alternatives?scryfall_id=...`. Render them as an image grid.

3. **Upload:** Form submit reads file input + selected `scryfall_id`, sends `POST /api/alternatives` as `FormData`. On success, re-fetch alternatives to show the new image.

### Styling (inline CSS)

- Dark theme (card images look best on dark backgrounds)
- Responsive grid for card images: `display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));`
- Cards have subtle hover effect (scale/border highlight)
- Selected card has a visible border/glow
- Upload form only visible when a card is selected
- Loading spinner or "Searching..." / "Loading..." text during fetches

## Step 8 — Error Handling & Polish

- Frontend: show error messages for failed searches, failed uploads (e.g., network error, file too large, wrong type).
- Backend: log errors, return proper HTTP status codes, sanitize user input (reject non-image uploads, limit filename length).
- Rate limiting: Scryfall asks for 50-100ms between requests. For MVP, no parallel requests — serialize searches. Add a comment noting this for future improvement.

## Step 9 — Testing

Manual testing checklist:

1. Start server, open in browser.
2. Search for a card (e.g., "Black Lotus") — verify images appear.
3. Click a card — verify detail section shows, alternatives section loads (empty initially).
4. Upload an image for the selected card — verify it appears in the alternatives gallery.
5. Refresh page, search again, select the same card — verify previously uploaded alternative persists.
6. Test edge cases: empty search, no results, oversized file, wrong file type.

---

## Implementation Order

Follow these steps sequentially. Each step produces a runnable (or at least compilable) state:

1. **Step 1** — Project skeleton, module, dependencies
2. **Step 2** — Database layer (can be tested independently)
3. **Step 5** — Server scaffold with routes returning placeholder responses
4. **Step 6** — Entry point wiring, verify server starts
5. **Step 3** — Scryfall search proxy
6. **Step 4** — Alternatives API
7. **Step 7** — Frontend HTML/CSS/JS
8. **Step 8** — Error handling polish
9. **Step 9** — Manual testing
