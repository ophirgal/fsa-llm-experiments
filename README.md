# FSA Boilerplate

Full-stack application boilerplate for AI-assisted coding interviews (60–90 min).

**Stack:** Go + Gin · React + TypeScript + TailwindCSS + shadcn/ui + Axios + React Query · PostgreSQL · Nginx · Docker Compose

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2
- `make`

---

## Quick Start

```bash
cp .env.example .env
make up
```

- App: http://localhost:3000
- API health: http://localhost:3000/api/v1/health

---

## Architecture

```
Browser → nginx:3000
              ├── /api/v1/*  → backend:8080  (Go + Gin, hot reload via Air)
              └── /*         → frontend:5173 (React + Vite, HMR)
                                    │
                               postgres:5432
```

All traffic flows through a single Nginx entry point.

---

## Development

| Command | Description |
|---|---|
| `make up` | Build and start all services |
| `make down` | Stop all services |
| `make logs` | Tail logs from all services |
| `make psql` | Open a psql shell in the database |
| `make clean` | Stop services and wipe volumes |

Both the backend (Air) and frontend (Vite) support hot reload — file saves are reflected immediately without restarting containers.

---

## Project Structure

```
fsa-boilerplate/
├── backend/
│   ├── api/             # HTTP handlers and router (/api/v1 routes)
│   ├── dal/             # data access layer (DB connection, migrations)
│   ├── config/          # environment configuration
│   └── migrations/      # *.sql files run in order at startup
├── frontend/
│   ├── components.json  # shadcn/ui CLI config
│   └── src/
│       ├── api/         # Axios client (base URL: /api/v1)
│       ├── hooks/       # React Query custom hooks
│       ├── lib/         # utilities (cn(), etc.)
│       ├── components/
│       │   └── ui/      # shadcn/ui components
│       └── pages/       # route-level page components
└── nginx/
    ├── nginx.conf       # dev proxy config
    └── nginx.prod.conf  # prod static file server + proxy
```

---

## Adding a New Resource

**1. Migration** — `backend/migrations/002_<name>.sql`
```sql
CREATE TABLE IF NOT EXISTS items (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**2. Handler** — add to `backend/api/handlers.go`
```go
func ListItems(c *gin.Context) { ... }
func CreateItem(c *gin.Context) { ... }
```

**3. Route** — add to `backend/api/router.go` in the `v1` group
```go
v1.GET("/items", ListItems)
v1.POST("/items", CreateItem)
```

**4. Hook** — `frontend/src/hooks/useItems.ts` with React Query
```ts
export function useItems() {
  return useQuery({
    queryKey: ['items'],
    queryFn: () => api.get<Item[]>('/items').then(res => res.data)
  })
}
```

**5. Page** — `frontend/src/pages/ItemsPage.tsx` using the hook

---

## shadcn/ui

UI components live in `frontend/src/components/ui/`. A `Button` component is included as a starter. Add more from the shadcn registry:

```bash
cd frontend
npx shadcn@latest add card
npx shadcn@latest add input
npx shadcn@latest add dialog
```

Import components using the `@` alias:

```tsx
import { Button } from '@/components/ui/button'

<Button variant="outline">Click me</Button>
```

Use `cn()` to merge Tailwind classes conditionally:

```tsx
import { cn } from '@/lib/utils'

<div className={cn('p-4', isActive && 'bg-primary text-primary-foreground')} />
```

Theme tokens (`bg-primary`, `text-muted-foreground`, etc.) are defined as CSS variables in `src/index.css` and mapped in `tailwind.config.js`.

---

## Production

```bash
make build-prod   # builds static frontend, runs backend binary
```

In production, Nginx serves the compiled React assets from `dist/` and proxies `/api/v1/*` to the Go backend. No Vite or Node process runs.
