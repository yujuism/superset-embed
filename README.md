# Superset Embed

Embedded analytics using Apache Superset inside a SvelteKit frontend, secured by a Go backend.

## Stack

- **Go** — guest token proxy, future RBAC/RLS
- **SvelteKit** — frontend with embedded dashboard
- **Apache Superset** — dashboard engine
- **PostgreSQL** — sales data + Superset metadata
- **Redis** — Superset cache

## How it works

```
Browser (Svelte)
  │
  │  POST /api/guest-token
  ▼
Go backend  ──────────────────►  Superset API
  │          admin credentials    returns JWT
  │
  │  embedDashboard(jwt)
  ▼
Superset iframe  ──────────────►  PostgreSQL
                    live query     sales data
```

The Go backend is the only service that holds Superset admin credentials. It issues short-lived guest JWTs to the frontend. RLS rules in the token control what data each user can see.

## Setup

### 1. Start services

```bash
docker compose build superset   # first time only — installs psycopg2
docker compose up -d
```

### 2. Connect Superset to the sales database

Open http://localhost:8088 → `admin / admin`

**Settings → Database Connections → + Database**

| Field | Value |
|---|---|
| Engine | PostgreSQL |
| Host | `data-db` |
| Port | `5432` |
| Database | `salesdb` |
| Username | `salesuser` |
| Password | `salespass` |

### 3. Create a dataset and dashboard

- **Datasets → + Dataset** → pick `salesdb` → `public` → `sales`
- Build charts, save to a Dashboard
- Dashboard → `···` → **Embed dashboard**
  - Allowed Domains: `http://localhost:5173`
  - Copy the UUID

### 4. Configure the frontend

```bash
cp frontend/.env.example frontend/.env
```

```env
VITE_BACKEND_URL=http://localhost:3000
VITE_SUPERSET_URL=http://localhost:8088
VITE_DASHBOARD_ID=<uuid-from-embed-dialog>
```

### 5. Run the frontend

```bash
cd frontend
pnpm install
pnpm dev
```

Open http://localhost:5173.

## Services

| Service | Port |
|---|---|
| Superset | 8088 |
| Go backend | 3000 |
| Svelte frontend | 5173 |
| PostgreSQL (sales data) | 5433 |
