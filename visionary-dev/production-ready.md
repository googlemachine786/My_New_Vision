# 🚀 AI Education Platform – System Architecture (MVP)

## 📌 Overview

This document defines the **full system architecture** for our AI-powered education platform.

The system is designed to support:

- 10k–20k concurrent users
- Real-time AI chat (streaming like ChatGPT)
- Multi-agent AI responses (text, diagram, image, video)
- RAG (PDF + document understanding)
- Subscription-based usage

---

# 🧱 Tech Stack

## Frontend
- Next.js (App Router)
- TailwindCSS + UI library
- Zustand (state management)
- React Query (API caching)
- SSE (Server-Sent Events) for streaming

## Backend
- NestJS (modular architecture)
- PostgreSQL (via Supabase)
- Redis (rate limit + queue + cache)
- BullMQ (job queue)

## AI & Data
- LLM APIs (OpenAI / Claude / Gemini)
- Vector DB (Supabase pgvector)
- Object Storage (Supabase Storage / S3)

## Infra
- Load Balancer (NGINX / Cloudflare)
- CDN (Cloudflare)
- Monitoring (Sentry / Logs)

---

# 🧠 High-Level Architecture
User (Browser)
│
▼
Frontend (Next.js)
│
▼
API Gateway (NestJS)
│
▼
Orchestrator Service
│
├── AI Agents
│ ├── Content Agent
│ ├── Diagram Agent
│ ├── Image Agent
│ ├── Resource Agent
│
▼
AI Services (LLM APIs)
│
▼
Data Layer
├── PostgreSQL (Supabase)
├── Vector DB (pgvector)
├── Redis
└── Storage (S3 / Supabase)

---

# 🔁 Full Request Workflow

## 1️⃣ User Sends Prompt

Frontend:

- Add user message instantly
- Create empty assistant message
- Start streaming request

---

# 🔁 Full Request Workflow

## 1️⃣ User Sends Prompt

Frontend:

- Add user message instantly
- Create empty assistant message
- Start streaming request
User → "Explain photosynthesis with diagram"


---

## 2️⃣ API Gateway (NestJS)

Handles:

- JWT Authentication
- Rate limiting (Redis)
- Plan validation
- Request logging

---

## 3️⃣ Orchestrator Service

Analyzes user intent:
{
explanation: true,
diagram: true,
video: false
}


Then triggers agents.

---

## 4️⃣ Multi-Agent Execution

### Content Agent
- Generates explanation via LLM

### Diagram Agent
- Generates Mermaid / SVG diagram

### Image Agent
- Calls image generation API

### Resource Agent
- Fetches YouTube / learning content

---

## 5️⃣ Streaming Response (SSE)

Backend streams tokens:
"H"
"He"
"Hel"
"Hello"


Frontend updates UI in real-time.

---

## 6️⃣ Save Data

- Store conversation in PostgreSQL
- Update usage limits
- Cache frequent responses (Redis)

---

# ⚡ Streaming Architecture
Frontend
│
▼
SSE Connection
│
▼
NestJS Controller
│
▼
LLM Streaming API
│
▼
Token chunks → Frontend UI

---

# 📦 Async Job Flow (Heavy Tasks)

Used for:

- Image generation
- PDF processing
- Video generation
User Request
│
▼
Queue (BullMQ)
│
▼
Worker Server
│
▼
Process AI Task
│
▼
Store Result (S3)
│
▼
Return Result URL

---

# 📚 RAG (Document AI System)

## Upload Flow
PDF Upload
↓
Text Extraction
↓
Chunking
↓
Embeddings
↓
Store in Vector DB

## Query Flow
User Question
↓
Vector Search
↓
Top Chunks
↓
LLM Response

---

# 🔐 Authentication & Plans

## Auth State
user
token
plan
limits
isAuthenticated


## Plan Example

### Free
- 10 prompts/day
- 3 images
- 1 PDF

### Pro
- 200 prompts/day
- 50 images
- 10 PDFs

---

# 💳 Subscription Flow
User Click Upgrade
↓
Create Payment Session
↓
Payment Gateway (Stripe/Razorpay)
↓
Webhook → Backend
↓
Update Plan in DB
↓
Frontend Refresh

---

# 🧠 Frontend Architecture

## State Management (Zustand)

### authStore

user
token
plan
limits


### chatStore
messages[]
isStreaming


### jobStore
jobId
status
resultUrl

### uiStore
modalOpen
sidebarOpen
theme

---

# 🎨 Frontend Modules

- Chat UI (streaming)
- Sidebar (history)
- Upload (PDF/video)
- Subscription UI
- Dashboard

---

# ⚙️ Backend Modules (NestJS)
auth/
chat/
agents/
orchestrator/
queue/
redis/
prisma/
payments/

---

# 📊 Scaling Architecture (10k–20k Users)

## API Layer

- 3–6 NestJS instances
- Stateless

## Worker Layer

- 2+ worker servers

## Database

- PostgreSQL (Supabase)
- Connection pooling

## Cache

- Redis

## Storage

- S3 / Supabase

---

# 🔍 Monitoring & Logs

Track:

- Errors (Sentry)
- API latency
- AI cost
- Token usage
- Queue jobs

---

# 🧩 Final System Layers
Frontend Layer (Next.js)
API Layer (NestJS)
Orchestrator Layer
AI Agent Layer
Data Layer (DB + Vector + Storage)
Infra Layer (Redis + Queue + CDN)

---

# 🚀 Summary

This architecture ensures:

- Real-time AI streaming
- Scalable multi-agent system
- Efficient cost control
- Clean separation of concerns
- Ready for 20k+ users

---

# 📌 Next Steps

- Integrate Figma UI into Next.js
- Implement Zustand stores
- Setup NestJS modules
- Configure Supabase (DB + Storage)
- Add Redis + Queue system
- Connect AI APIs

---
