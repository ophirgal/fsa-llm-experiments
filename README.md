# LLM Experiment Runner

A tool for evaluating LLM system prompt variants against a dataset of queries using an LLM-as-judge scoring approach.

**Stack:** Go + Gin · React + TypeScript + TailwindCSS + shadcn/ui + Axios + React Query · PostgreSQL · Nginx · Docker Compose

---

## What It Does

1. **Upload a dataset** — upload a plain text file where each line is a query
2. **Create an experiment** — pick a dataset, write one or more system prompt variants, and write a shared judge prompt
3. **Run the experiment** — each query is sent to the LLM once per variant; every response is then scored automatically by the judge prompt (1 = pass, 0 = fail)
4. **View results** — see a per-variant pass rate: how many queries scored 1 out of the total

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
fsa-llm-experiments/
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

**1. Migration** — `backend/migrations/NNN_<name>.sql`
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

## Core Domain

| Entity | Description |
|---|---|
| **Dataset** | A named collection of queries parsed from an uploaded text file (one query per line) |
| **Experiment** | Ties a dataset to one or more system prompt variants and a shared judge prompt |
| **Run** | One execution of an experiment — sends each query through the LLM for each variant |
| **Result** | The LLM response + judge score (1/0) for a single query × variant pair |

---

## shadcn/ui

UI components live in `frontend/src/components/ui/`. Add more from the shadcn registry:

```bash
cd frontend
npx shadcn@latest add card
npx shadcn@latest add input
npx shadcn@latest add dialog
```

Import with the `@` alias:

```tsx
import { Button } from '@/components/ui/button'
```

Use `cn()` to merge Tailwind classes:

```tsx
import { cn } from '@/lib/utils'
<div className={cn('p-4', isActive && 'bg-primary text-primary-foreground')} />
```
