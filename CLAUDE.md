# Project Context

This is a full-stack web application built for a coding interview. The goal is to implement the required features as quickly and cleanly as possible.

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.23 + Gin |
| Database | PostgreSQL 16 (via `database/sql` + pgx driver) |
| Frontend | React 18 + TypeScript + TailwindCSS 3 + shadcn/ui + Axios + React Query |
| Dev server | Vite 5 (HMR enabled) |
| Reverse proxy | Nginx (single entry point) |
| Orchestration | Docker Compose |

## Running the Project

```bash
cp .env.example .env   # first time only
make up                # build + start all services
make logs              # watch logs
make psql              # database shell
make down              # stop
make clean             # stop + wipe volumes
```

App is at **http://localhost:3000**. API is at **http://localhost:3000/api/v1/**.

## Architecture

```
Browser → nginx:80 (host: 3000)
              ├── /api/v1/*  → backend:8080
              └── /*         → frontend:5173 (dev) / static files (prod)
```

Both backend (Air) and frontend (Vite) hot-reload on file save. No container restarts needed during development.

## Key Files

| File | Purpose |
|---|---|
| `backend/main.go` | Entry point: load config, connect DB, run migrations, start server |
| `backend/api/router.go` | Route registrations in `/api/v1` group |
| `backend/api/handlers.go` | HTTP handlers (add more files per resource as needed) |
| `backend/dal/db.go` | `NewDB()` and `RunMigrations()` |
| `backend/config/config.go` | Configuration loading |
| `backend/migrations/` | SQL files run in alphabetical order at startup |
| `frontend/src/api/client.ts` | Axios instance with base URL `/api/v1` |
| `frontend/src/hooks/` | React Query custom hooks for API calls |
| `frontend/src/pages/` | Top-level page components |
| `frontend/src/components/` | Reusable UI components (shadcn/ui components live in `components/ui/`) |
| `frontend/src/lib/utils.ts` | `cn()` utility for merging Tailwind classes |
| `frontend/components.json` | shadcn/ui CLI config |

## Adding a New Resource (End-to-End)

### 1. Migration
Create `backend/migrations/NNN_<name>.sql` (NNN = next number):
```sql
CREATE TABLE IF NOT EXISTS items (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
Migrations run automatically on next backend start.

### 2. Handler
Add to `backend/api/handlers.go` (or create separate file if handler is complex):
```go
func ListItems(c *gin.Context) {
    rows, err := db.QueryContext(c.Request.Context(), `SELECT id, name, created_at FROM items ORDER BY id`)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    // scan rows...
    c.JSON(http.StatusOK, items)
}
```

### 3. Route
In `backend/api/router.go`, add inside the `v1` group:
```go
v1.GET("/items", ListItems)
v1.POST("/items", CreateItem)
```

### 4. Frontend Hook
Create `frontend/src/hooks/useItems.ts`:
```ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../api/client'

export function useItems() {
  return useQuery({
    queryKey: ['items'],
    queryFn: () => api.get<Item[]>('/items').then(res => res.data)
  })
}

export function useCreateItem() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateItemInput) => api.post('/items', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['items'] })
  })
}
```

Then use in components:
```ts
const { data: items, isLoading } = useItems()
const { mutate: createItem } = useCreateItem()
```

### 5. Page
Create `frontend/src/pages/ItemsPage.tsx` and add it to the router in `App.tsx`.

## Database Conventions

- All tables use `BIGSERIAL` primary keys named `id`
- Timestamps use `TIMESTAMPTZ` with `DEFAULT NOW()`
- Migration files: `NNN_description.sql` (e.g., `002_add_items.sql`)
- Use `IF NOT EXISTS` guards so migrations are idempotent

## Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | Full postgres connection string |
| `PORT` | Backend port (default: 8080) |
| `POSTGRES_USER` | DB user (used by postgres container) |
| `POSTGRES_PASSWORD` | DB password |
| `POSTGRES_DB` | DB name |

## Go Module

Module name: `fsa-boilerplate/backend`

Import paths follow `fsa-boilerplate/backend/<package>` (e.g., `fsa-boilerplate/backend/api`, `fsa-boilerplate/backend/dal`, `fsa-boilerplate/backend/config`).

## Tailwind & shadcn/ui

Tailwind classes are available everywhere in `src/`. shadcn/ui components live in `src/components/ui/` and use CSS variable–based tokens (`bg-primary`, `text-muted-foreground`, etc.) defined in `src/index.css`.

The `@` alias resolves to `src/`, so imports look like `import { Button } from '@/components/ui/button'`.

**Adding a shadcn component** (run from `frontend/`):
```bash
npx shadcn@latest add <component>   # e.g. npx shadcn@latest add card
```
This drops the component file into `src/components/ui/` and installs any missing Radix peer deps.

**`cn()` utility** — merge Tailwind classes safely:
```ts
import { cn } from '@/lib/utils'
<div className={cn('px-4', isActive && 'bg-primary')} />
```

Common token classes:
- Backgrounds: `bg-background`, `bg-card`, `bg-primary`, `bg-muted`
- Text: `text-foreground`, `text-muted-foreground`, `text-primary-foreground`
- Borders: `border-border`, `border-input`
- Interactive: `hover:bg-accent`, `focus-visible:ring-2 focus-visible:ring-ring`

## Autonomy
- When asked to implement a plan, execute it fully end-to-end — including running tests and verification — without asking for permission at each step.

## General Coding Guidelines
- When adding code for a new feature, make minimal changes needed to implement the feature while still sticking to good practices (e.g. do not hardcode things just because it's more minimal). 
- Do not overengineer and add more files than are needed for the feature.
- For backend features, write tests first and use those to validate the feature's implementation (TDD). One happy path test and one error case test is good enough, no need to go crazy with tests. But you may add more tests if you think they are important.
- For frontend tests, use jest and the React Testing library to test with user events and by using jest to mock API calls.
