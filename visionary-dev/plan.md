# Visionary MVP – Step-by-Step Build Plan

> Deadline: May 15, 2026 | Payments/Subscriptions: SKIPPED for MVP

---

## How to Use This Plan

1. Every time a new Figma design is added to `figma_design_1/` (or new folders), analyze it and update this plan.
2. Work feature by feature, screen by screen.
3. Cross-reference `workflow.md` for how each feature connects end-to-end.
4. All items marked **[MANUAL CONFIG]** require you to set up keys, services, or configs manually.

---

## Phase 0 – Project Setup (Day 1–2)

### Frontend (Next.js)
- [ ] Init Next.js 14+ with App Router
- [ ] Install: TailwindCSS, shadcn/ui, Zustand, React Query, axios
- [ ] Setup folder structure (see `/frontend/`)
- [ ] Configure environment variables

**[MANUAL CONFIG] Frontend `.env.local`:**
```
NEXT_PUBLIC_API_URL=http://localhost:3001
NEXT_PUBLIC_SUPABASE_URL=your_supabase_url
NEXT_PUBLIC_SUPABASE_ANON_KEY=your_supabase_anon_key
```

### Backend (NestJS)
- [ ] Init NestJS project
- [ ] Install: Prisma, Passport JWT, Redis (ioredis), BullMQ, class-validator
- [ ] Setup folder structure (see `/backend/`)
- [ ] Configure environment variables

**[MANUAL CONFIG] Backend `.env`:**
```
DATABASE_URL=postgresql://...@supabase...
JWT_SECRET=your_super_secret_key
REDIS_URL=redis://localhost:6379
OPENAI_API_KEY=sk-...
CLAUDE_API_KEY=sk-ant-...
SUPABASE_URL=your_supabase_url
SUPABASE_SERVICE_KEY=your_service_role_key
SUPABASE_STORAGE_BUCKET=uploads
```

**[MANUAL CONFIG] External Services to Set Up:**
- [ ] Create Supabase project → get DB URL, anon key, service key
- [ ] Enable pgvector extension in Supabase SQL editor: `CREATE EXTENSION vector;`
- [ ] Create Redis instance (local Docker or Upstash for prod)
- [ ] Get OpenAI / Claude API key
- [ ] Create Supabase Storage bucket named `uploads`

---

## Phase 1 – Authentication (Week 1)

> Depends on: Figma auth screens (Login, Signup, Onboarding)

### Backend
- [ ] `auth` module: register, login, JWT strategy
- [ ] Prisma: `User` table migration
- [ ] Google OAuth (Passport strategy)
- [ ] JWT refresh token logic

### Frontend
- [ ] `/login` page (Figma-driven)
- [ ] `/signup` page (Figma-driven)
- [ ] `authStore` (Zustand): `user`, `token`, `plan`, `isAuthenticated`
- [ ] Axios interceptor with JWT Bearer header
- [ ] Redirect logic post-login → `/dashboard` or `/chat`

**[MANUAL CONFIG] Google OAuth:**
- Create Google Cloud project
- Enable Google+ API
- Create OAuth 2.0 credentials
- Add to backend `.env`: `GOOGLE_CLIENT_ID=...` `GOOGLE_CLIENT_SECRET=...` `GOOGLE_CALLBACK_URL=http://localhost:3001/auth/google/callback`

---

## Phase 2 – Dashboard & Layout (Week 1–2)

> Depends on: Figma dashboard/home screen

- [ ] Root layout with sidebar + main content area
- [ ] Sidebar: chat history list, navigation links
- [ ] Dashboard home page (recent chats, quick start)
- [ ] `uiStore` (Zustand): `sidebarOpen`, `theme`, `modalOpen`
- [ ] Responsive layout (mobile-first from Figma)

---

## Phase 3 – Chat UI & Streaming (Week 2–3)

> Depends on: Figma chat screen

### Frontend
- [ ] `/chat/[id]` page
- [ ] Message list component (user + assistant bubbles)
- [ ] Prompt input with send button
- [ ] Streaming via SSE: token-by-token updates
- [ ] `chatStore` (Zustand): `messages[]`, `isStreaming`, `currentChatId`
- [ ] Empty assistant message bubble → fills as tokens arrive
- [ ] Markdown rendering in assistant messages
- [ ] Code block syntax highlighting

### Backend
- [ ] `chat` module: create conversation, get history
- [ ] SSE endpoint: `GET /chat/stream`
- [ ] LLM integration (OpenAI/Claude streaming)
- [ ] Save conversation + messages to PostgreSQL

---

## Phase 4 – Multi-Agent Orchestrator (Week 3–4)

> Depends on: Phase 3 complete

### Backend
- [ ] `orchestrator` module: intent detection
- [ ] Intent schema: `{ explanation, diagram, image, resource }`
- [ ] `content-agent`: LLM text generation
- [ ] `diagram-agent`: Mermaid/SVG generation
- [ ] `resource-agent`: YouTube/web resource fetching

### Frontend
- [ ] Render diagram blocks (Mermaid renderer)
- [ ] Render resource cards (YouTube embeds/links)
- [ ] Multi-part response display (text + diagram + resources together)

---

## Phase 5 – PDF Upload & RAG (Week 4–5)

> Depends on: Figma upload screen

### Backend
- [ ] File upload endpoint → Supabase Storage
- [ ] PDF text extraction (pdf-parse or pdfjs)
- [ ] Chunking + embedding generation (OpenAI embeddings)
- [ ] Store chunks in pgvector (`document_chunks` table)
- [ ] RAG query: vector search → top chunks → LLM response

### Frontend
- [ ] Upload component (drag-drop or click)
- [ ] Upload progress indicator
- [ ] "Chat with PDF" mode UI
- [ ] Show source citations in responses

**[MANUAL CONFIG] pgvector:**
```sql
-- Run in Supabase SQL editor
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE document_chunks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id uuid REFERENCES documents(id),
  content text,
  embedding vector(1536),
  created_at timestamptz DEFAULT now()
);
CREATE INDEX ON document_chunks USING ivfflat (embedding vector_cosine_ops);
```

---

## Phase 6 – Async Jobs (Image Generation) (Week 5)

> Depends on: BullMQ + Redis setup

### Backend
- [ ] `queue` module with BullMQ
- [ ] Image generation worker
- [ ] Job status endpoint: `GET /jobs/:id`
- [ ] Store result URL in S3/Supabase Storage

### Frontend
- [ ] `jobStore` (Zustand): `jobId`, `status`, `resultUrl`
- [ ] Poll job status → show progress
- [ ] Display generated image in chat

---

## Phase 7 – Polish & Testing (Week 6 → May 15)

- [ ] Error boundaries + toast notifications
- [ ] Loading states across all features
- [ ] Mobile responsiveness (match Figma exactly)
- [ ] API error handling
- [ ] Basic E2E test for auth + chat flow
- [ ] Deploy backend (Railway/Render)
- [ ] Deploy frontend (Vercel)
- [ ] Point to production Supabase

**[MANUAL CONFIG] Production Deploy:**
- Vercel: set all `NEXT_PUBLIC_*` env vars in dashboard
- Railway/Render: set all backend env vars
- Update `NEXT_PUBLIC_API_URL` to production backend URL
- Update Google OAuth callback URL to production domain
- Enable CORS on backend for production frontend domain

---

## Database Tables (Prisma Migrations)

| Table | Purpose |
|---|---|
| `users` | Auth, profile, plan info |
| `conversations` | Chat sessions per user |
| `messages` | Individual chat messages |
| `documents` | Uploaded PDF metadata |
| `document_chunks` | RAG vector chunks (pgvector) |
| `jobs` | Async job tracking (image gen, etc.) |

Full schema in `backend/prisma/schema.prisma`.

---

## Figma Design Integration Process

Every time a new Figma design is added to the project:
1. Drop PNG/JPG exports into `figma_design_1/` (or new `figma_design_N/` folder)
2. Tell Claude: "New designs added, analyze and proceed"
3. Claude will read all images in the folder, identify the screen, and implement that feature next
4. Update this plan file with checkboxes for the new screen
5. Update `workflow.md` with the UX flow

---

## Current Status

- [x] Architecture defined (`production-ready.md`)
- [x] Base folder structures created
- [ ] Phase 0: Project setup
- [ ] Phase 1: Authentication
- [ ] Phase 2: Dashboard layout
- [ ] Phase 3: Chat + streaming
- [ ] Phase 4: Multi-agent
- [ ] Phase 5: RAG
- [ ] Phase 6: Async jobs
- [ ] Phase 7: Polish + deploy
