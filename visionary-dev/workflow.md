# Visionary MVP – Feature Workflows

> This file describes how every feature works end-to-end in the MVP.
> Updated each time new Figma designs are analyzed.

---

## 1. Authentication Flow

```
User visits app
  → Not logged in → redirect to /login
  → Login page (Figma screen)
  → Enter email + password OR click "Continue with Google"

Email/Password path:
  Frontend → POST /auth/login { email, password }
  Backend → validate credentials → return { accessToken, user }
  Frontend → authStore.setUser(user) + authStore.setToken(token)
  Frontend → store token in localStorage
  Frontend → redirect to /dashboard

Google OAuth path:
  Frontend → GET /auth/google (backend redirect)
  Backend → Google OAuth → callback → issue JWT
  Frontend → receive token → same as above

On every request:
  Axios interceptor → adds Authorization: Bearer <token>
  Backend JWT Guard → validates token → attaches user to request

Logout:
  authStore.clear() → remove localStorage token → redirect /login
```

---

## 2. Dashboard Flow

```
User arrives at /dashboard
  → Sidebar loads: recent conversations from GET /conversations
  → Main area: welcome screen OR last conversation
  → Click "New Chat" → navigate to /chat/new
  → Click existing chat → navigate to /chat/:id → load history
```

---

## 3. Chat & Streaming Flow

```
User types prompt → clicks Send (or hits Enter)

Frontend:
  1. chatStore.addMessage({ role: 'user', content: prompt })
  2. chatStore.addMessage({ role: 'assistant', content: '' })  ← empty bubble
  3. chatStore.setStreaming(true)
  4. Open SSE connection: GET /chat/stream?prompt=...&conversationId=...

Backend SSE endpoint:
  1. Validate JWT
  2. Save user message to DB
  3. Call orchestrator → detect intent { explanation, diagram, image, resource }
  4. Stream LLM tokens → send as SSE events: data: {"token": "H"}
  5. On finish → send: data: {"done": true, "messageId": "..."}
  6. Save full assistant message to DB
  7. Update usage counter

Frontend (SSE handler):
  - On each token: chatStore.appendToken(token) → updates last message content
  - On done: chatStore.setStreaming(false) → finalize message
  - Render markdown in real-time as tokens arrive

Visual: User sees text appearing word-by-word like ChatGPT
```

---

## 4. Multi-Agent Orchestrator Flow

```
User prompt: "Explain photosynthesis with a diagram"

Orchestrator:
  → LLM call to classify intent
  → Returns: { explanation: true, diagram: true, image: false, resource: true }

Parallel agent execution:
  ┌─ Content Agent → LLM streaming → text explanation
  ├─ Diagram Agent → LLM → Mermaid code → rendered diagram
  └─ Resource Agent → YouTube API / web search → relevant links

Response structure:
  {
    text: "Photosynthesis is...",
    diagram: "graph TD\n  Sun-->Plant...",
    resources: [{ title: "...", url: "..." }]
  }

Frontend renders:
  - Text (streamed, markdown)
  - Mermaid diagram (rendered client-side)
  - Resource cards (link + thumbnail)
```

---

## 5. PDF Upload & RAG Flow

### Upload:
```
User clicks upload → selects PDF

Frontend:
  → POST /documents/upload (multipart form)
  → Show upload progress bar
  → On success: show "PDF ready, ask anything!"

Backend:
  → Save file to Supabase Storage
  → Extract text with pdf-parse
  → Split into 500-token chunks with 50-token overlap
  → Generate embeddings for each chunk (OpenAI text-embedding-3-small)
  → Store chunks + embeddings in document_chunks (pgvector)
  → Save document metadata to documents table
```

### Query (RAG Chat):
```
User asks: "What is the conclusion of this paper?"

Backend:
  → Generate embedding for question
  → pgvector similarity search → top 5 chunks
  → Build context: "Using these excerpts: [chunk1][chunk2]..."
  → LLM call with context → generate answer
  → Stream answer back via SSE

Frontend:
  → Shows answer with source citations
  → "From page 3: ..." reference cards below answer
```

---

## 6. Async Job Flow (Image Generation)

```
User asks for an image: "Generate a diagram of the solar system"

Frontend:
  → POST /jobs/image { prompt }
  → Receive: { jobId: "abc123" }
  → jobStore.setJob({ jobId, status: 'pending' })
  → Show spinner: "Generating image..."

Backend:
  → Add job to BullMQ queue
  → Return jobId immediately (non-blocking)
  → Worker picks up job → calls image generation API
  → Saves result image to Supabase Storage
  → Updates job status to 'completed' + stores resultUrl

Frontend polling:
  → GET /jobs/abc123 every 3 seconds
  → On status = 'completed': jobStore.setResult(resultUrl)
  → Display image in chat

Error handling:
  → On status = 'failed': show error toast
  → Allow user to retry
```

---

## 7. Sidebar & Conversation History

```
On load:
  GET /conversations → list of user's past conversations
  → Render in sidebar: title (first message) + date

On select:
  GET /conversations/:id/messages → full message history
  → chatStore.setMessages(messages)
  → Render full conversation

On new chat:
  POST /conversations → create new
  → Navigate to /chat/:newId
  → chatStore.reset()

Auto-title:
  After first assistant response → backend generates a short title
  → sidebar updates with title
```

---

## 8. State Management Summary (Zustand)

### authStore
```ts
{
  user: User | null
  token: string | null
  isAuthenticated: boolean
  setUser(user): void
  setToken(token): void
  clear(): void
}
```

### chatStore
```ts
{
  messages: Message[]
  isStreaming: boolean
  currentConversationId: string | null
  addMessage(msg): void
  appendToken(token): void   // appends to last assistant message
  setStreaming(val): void
  setMessages(msgs): void
  reset(): void
}
```

### jobStore
```ts
{
  jobId: string | null
  status: 'idle' | 'pending' | 'completed' | 'failed'
  resultUrl: string | null
  setJob(jobId, status): void
  setResult(url): void
  reset(): void
}
```

### uiStore
```ts
{
  sidebarOpen: boolean
  modalOpen: boolean
  theme: 'light' | 'dark'
  toggleSidebar(): void
  setModal(open): void
  setTheme(theme): void
}
```

---

## Figma Design → Workflow Update Process

When new Figma designs are added:
1. Identify the screen name (e.g., "Chat Screen", "Upload Modal")
2. Add a new section to this file describing the UX flow
3. Map each UI element to the corresponding API call and store action
4. Update `plan.md` with implementation tasks for that screen
