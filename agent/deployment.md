## Step 10 — Deployment

Deploy to a free-tier cloud platform. Recommended: **Fly.io** (free tier includes a small VM, persistent volume for SQLite + uploads).

### Deployment Steps (Fly.io)

1. **Install Fly CLI:** `curl -L https://fly.io/install.sh | sh`
2. **Authenticate:** `fly auth login`
3. **Create app:** `fly launch` — choose nearest region, no database addon needed
4. **Configure persistent volume** (so SQLite DB and uploads survive deploys):
   - `fly volumes create data --size 1`
5. **Create `fly.toml`** with:
   ```
   [build]
     Dockerfile = "Dockerfile"

   [mounts]
     source = "data"
     destination = "/data"

   [env]
     PORT = "8080"
     DB_PATH = "/data/data.db"
     UPLOADS_DIR = "/data/uploads"
   ```
6. **Create `Dockerfile`:**
   ```dockerfile
   FROM golang:1.23-alpine AS build
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN CGO_ENABLED=1 go build -o /mtg-alternatives .

   FROM alpine:3.20
   RUN apk add --no-cache libstdc++
   COPY --from=build /mtg-alternatives /mtg-alternatives
   EXPOSE 8080
   CMD ["/mtg-alternatives"]
   ```
   Note: `CGO_ENABLED=1` is required for `go-sqlite3`.
7. **Deploy:** `fly deploy`
8. **Open app:** `fly open` — returns the public URL.
