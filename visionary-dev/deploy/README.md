# Visionary — Local Dev Infrastructure

Single source of truth for running Visionary's backend dependencies (Postgres, Redis, Mailpit) in containers while you develop the backend and frontend on the host with hot reload.

> Status: local development only. Staging and production are separate concerns and not yet defined here. See the team's planning docs for those.

---

## What this gives you

| Service | Purpose | Reachable from host |
|---|---|---|
| `postgres` | Application database | `localhost:54322` |
| `redis` | Cache + (future) BullMQ queue | `localhost:6379` |
| `mailpit` | Captures outgoing OTP emails locally — never sends real mail | SMTP `localhost:1025`, UI `http://localhost:8025` |
| `migrate` | One-shot init container — runs `prisma migrate deploy` and conditionally seeds master data, then exits | (no exposed port) |

The backend (`npm run start:dev` in `backend/`) and frontend (`npm run dev` in `frontend/`) run on the host, not in containers, so you keep hot reload.

### What about Supabase?

| Use | Where |
|---|---|
| **Object storage** (PDF uploads) | ✅ Cloud Supabase project (real internet call) |
| **Postgres database** | ❌ Not Supabase — local container in `compose.yml` |
| **Auth** | ❌ Not Supabase — NestJS owns auth (JWT + Google OAuth) |
| **Realtime / Edge / PostgREST** | ❌ Not used at all |
| **Supabase CLI / local Supabase stack** | ❌ Retired. Don't `supabase start` anymore |

So "Supabase" in this codebase = **cloud-hosted object storage only**. Everything else runs locally via this compose.

---

## Prerequisites

- **Docker** with Compose v2. [OrbStack](https://orbstack.dev) is the team's default on macOS — lighter than Docker Desktop and a drop-in replacement for the `docker` CLI.
- **Node 20.x** — match the version in `.nvmrc`. Use `nvm use` from the repo root.
- **A cloud Supabase project** for object storage — credentials live in the team secret store. Ask if you don't have access yet.

---

## First-time setup

```bash
# 1. Use Node 20
nvm use

# 2. Install backend deps (only needed once)
cd backend && npm install
cd ../frontend && npm install
cd ..

# 3. Copy env template and fill in cloud Supabase credentials
cp deploy/env/.env.example backend/.env
# Edit backend/.env:
#   - SUPABASE_URL          → from team's cloud Supabase project
#   - SUPABASE_SERVICE_KEY  → from team's cloud Supabase project
#   - SUPABASE_STORAGE_BUCKET → "uploads" (or whatever the team agreed on)
#   - GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET → from team
#   - OPENAI_API_KEY, CLAUDE_API_KEY        → from team (may be empty)
#   - JWT_SECRET            → generate fresh: openssl rand -hex 32

# 4. Frontend env
echo "NEXT_PUBLIC_API_URL=http://localhost:3001" > frontend/.env

# 5. Pull the pre-built migrate image from GHCR (first time only) + bring up the stack
docker compose -f deploy/compose.yml pull
docker compose -f deploy/compose.yml up -d
# If the GHCR pull fails (e.g. image not built yet), use the local-build fallback:
#   docker compose -f deploy/compose.yml up -d --build

# 6. Verify migrations + seed ran
docker compose -f deploy/compose.yml logs migrate
# Expect to see: "Running prisma migrate deploy..." then either
#   "Seeding master data..." (first time) or
#   "Master data already populated; skipping seed."

# 7. Start backend (terminal 1)
cd backend && nvm use && npm run start:dev

# 8. Start frontend (terminal 2)
cd frontend && nvm use && npm run dev

# Open http://localhost:3000
```

---

## Daily workflow

### Start everything

```bash
# Infra (already up if you didn't `down`)
docker compose -f deploy/compose.yml up -d

# Backend (terminal 1)
cd backend && nvm use && npm run start:dev

# Frontend (terminal 2)
cd frontend && nvm use && npm run dev
```

### Stop everything

```bash
# Backend / frontend: Ctrl+C in their terminals

# Infra: keeps running across reboots until you stop it.
docker compose -f deploy/compose.yml down
```

### Tail logs

```bash
docker compose -f deploy/compose.yml logs -f                    # all services
docker compose -f deploy/compose.yml logs -f postgres           # just one
docker compose -f deploy/compose.yml logs migrate               # init-container output
```

### Inspect the database

Two options:

```bash
# Prisma Studio — best for relational browsing
cd backend && npx prisma studio
# Opens http://localhost:5555

# psql shell
docker exec -it visionary-postgres psql -U postgres -d postgres
```

### Read captured emails (signup OTPs, etc.)

Open <http://localhost:8025> in a browser. Mailpit catches every email the backend sends.

### Inspect Redis

```bash
docker exec -it visionary-redis redis-cli
```

### Verify backend health

```bash
curl http://localhost:3001/api/health        # full check (DB + Redis)
curl http://localhost:3001/api/health/live   # liveness only — process alive
curl http://localhost:3001/api/health/ready  # readiness — DB + Redis up
```

---

## Storage configuration

The backend talks to **cloud Supabase Storage** for PDF uploads. There is no Storage container in this compose.

Why: cloud Supabase is what production uses, and emulating it locally adds a 12-container stack (Supabase CLI) that is 90% unused by Visionary's app code.

What this means for you:
- PDF upload features require an internet connection.
- Test uploads land in whatever bucket `SUPABASE_STORAGE_BUCKET` points at. Ideally the team has a separate non-prod bucket (e.g. `uploads-dev`) so test files don't mix with prod files. Confirm with the team.

---

## Re-importing prod data (optional)

When you want real-looking data in your local DB instead of just the seeded master data:

```bash
# 1. Get a fresh dump from cloud Supabase (one-time, requires DIRECT_URL credentials)
docker run --rm postgres:17 pg_dump --schema=public --no-owner --no-acl --clean --if-exists \
  "$DIRECT_URL_FROM_PROD" > local/prod-dump.sql

# 2. Wipe local public schema and restore
docker exec -i visionary-postgres psql -U postgres -d postgres <<EOF
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;
EOF
docker exec -i visionary-postgres psql -U postgres -d postgres < local/prod-dump.sql
```

> ⚠️ Prod dumps contain real user PII (emails, bcrypt hashes). Don't share screenshots, don't commit the `.sql` file (it's gitignored), and delete it when no longer needed.

---

## Common operations

### Re-run migrations after pulling new schema changes

The `migrate` init container only runs on `up`. To re-run after pulling schema changes:

```bash
docker compose -f deploy/compose.yml up -d --force-recreate migrate
```

Or run them directly from the backend (does the same thing, faster):

```bash
cd backend && nvm use && npx prisma migrate deploy
```

### Force a re-seed

The seed normally only runs against an empty `subject_master` table. To re-seed against a populated DB:

```bash
SEED_ON_BOOT=force docker compose -f deploy/compose.yml up -d --force-recreate migrate
docker compose -f deploy/compose.yml logs migrate
```

> ⚠️ The seed uses `deleteMany` on subjects/books/chapters. Because of `onDelete: Cascade` on `subject_track` and `chapter_track`, **a re-seed wipes any user progress in your local DB**. Don't force-seed if you have test users you care about.

### Wipe the database completely

```bash
docker compose -f deploy/compose.yml down -v   # -v also deletes volumes
docker compose -f deploy/compose.yml up -d --build
# migrate will run, see empty DB, and re-seed automatically
```

### Update the migrate image after a backend code change

The migrate image is built from `backend/`. After changing `prisma/seed.ts` or `prisma/schema.prisma`:

```bash
docker compose -f deploy/compose.yml build migrate
docker compose -f deploy/compose.yml up -d --force-recreate migrate
```

---

## File layout

```
deploy/
├── README.md                   ← this file
├── compose.yml                 ← service definitions (postgres, redis, mailpit, migrate)
├── Dockerfile.backend          ← multi-stage backend image
│                                  - target `runtime` (lean, prod build) — used by future staging
│                                  - target `migrate` (with dev deps for prisma + ts-node)
└── env/
    └── .env.example            ← template for backend/.env

backend/
├── scripts/
│   └── migrate-and-seed.sh     ← entrypoint for migrate init container
└── .dockerignore               ← excludes node_modules, .env*, dist, etc.
```

The backend's `src/health/` module exposes `/api/health`, `/api/health/live`, `/api/health/ready` — required for the Dockerfile's `HEALTHCHECK` and useful for any orchestrator down the line.

---

## Troubleshooting

### `EADDRINUSE: address already in use :::3001`

Another backend (or process) is on port 3001. Find and stop it:

```bash
lsof -nP -iTCP:3001 -sTCP:LISTEN
kill <PID>
```

If you previously ran the backend against the now-retired Supabase CLI, your old `npm run start:dev` may still be running.

### `Can't reach database server at 127.0.0.1:54322`

The Postgres container isn't running. Check OrbStack is running, then:

```bash
docker compose -f deploy/compose.yml up -d postgres
```

### Migrate container keeps restarting

It shouldn't — `restart: "no"` is set in `compose.yml`. If it does, check logs:

```bash
docker compose -f deploy/compose.yml logs migrate
```

Common causes:
- Postgres not yet healthy when migrate starts (rare; the `depends_on` health condition should prevent this)
- Schema drift between `schema.prisma` and the existing DB (delete the `postgres-data` volume and start over: `docker compose -f deploy/compose.yml down -v && up -d --build`)
- `prisma migrate deploy` requires migrations to exist in `backend/prisma/migrations/`. If you've squashed or deleted migrations, re-run `npx prisma migrate dev` against the local DB to create new ones.

### Backend can't talk to Mailpit

Confirm Mailpit is up: `curl -s http://localhost:8025 | head` should return HTML.
Confirm `backend/.env` has `SMTP_HOST=localhost SMTP_PORT=1025 SMTP_SECURE=false`.

### `prisma generate` fails inside the migrate build

Usually a Prisma binary target mismatch. If you see `Couldn't find Prisma engine for ...`, check `binaryTargets` in `schema.prisma`. The image uses `node:20-slim` (Debian, glibc) — `["native"]` should resolve to a Linux glibc binary.

### Volumes filling disk

```bash
docker system df                                # see usage
docker compose -f deploy/compose.yml down -v   # nuke this stack's volumes
docker volume prune                             # nuke unused volumes globally
```

---

## When to update what

| Change | Update |
|---|---|
| Added a new env var the backend reads | `deploy/env/.env.example` |
| Added a new service (e.g. ClickHouse, Elasticsearch) | `deploy/compose.yml` + this README |
| Bumped Postgres / Redis / Node major version | The `image:` tag in compose, and the `FROM` line in `Dockerfile.backend` |
| New Prisma migration | Just commit it to `backend/prisma/migrations/` — the init container picks it up on next `up` |
| Changed the seed script | Re-build the migrate image: `docker compose -f deploy/compose.yml build migrate` |

---

## Conventions

- **Branches**: `chore/<desc>`, `feature/<desc>`, `fix/<desc>`, `refactor/<desc>`, `tweak/<desc>`. One logical change per branch. See repo `CLAUDE.md` for the full policy.
- **The host runs the apps; the compose runs the deps.** Don't add `backend` or `frontend` services to this compose — staging compose will, this one shouldn't.
- **Secrets never get committed.** Anything matching `.env*` is gitignored. Use the team's secret store for actual values.
- **`postgres-data`, `redis-data`, `mailpit-data` are named Docker volumes.** They survive `docker compose down`. Use `down -v` only when you actually want to wipe.

---

## Related files (in the repo)

- [`backend/CLAUDE.md`-style files](../CLAUDE.md) — agent-facing rules and policies (gitignored on this branch; ask the team if you can't find it)
- [Project README.md](../README.md) — full architecture documentation; treat as authoritative for what's implemented
- [`backend/prisma/schema.prisma`](../backend/prisma/schema.prisma) — DB schema. If it disagrees with the prod DB, the prod DB has untracked drift (worth raising)
