# Visionary — Full Project Documentation


> AI-powered educational platform for students, teachers, and organizations.
> Built with **NestJS** (backend) · **Next.js 16** (frontend) · **PostgreSQL** · **OpenAI**

---

## Table of Contents

- [⚠️ CRITICAL — Read Before Making Any Changes](#️-critical--read-before-making-any-changes)

1. [Project Overview](#1-project-overview)
2. [Tech Stack](#2-tech-stack)
3. [Repository Structure](#3-repository-structure)
4. [Environment Setup](#4-environment-setup)
5. [Running the Project](#5-running-the-project)
6. [Frontend Architecture](#6-frontend-architecture)
7. [Design System](#7-design-system)
8. [Component Library](#8-component-library)
9. [State Management](#9-state-management)
10. [Routing & Middleware](#10-routing--middleware)
11. [Auth Flow](#11-auth-flow)
12. [Onboarding Flow](#12-onboarding-flow)
13. [Learning System](#13-learning-system)
14. [Profile & Settings](#14-profile--settings)
15. [Backend Architecture](#15-backend-architecture)
16. [Database Schema](#16-database-schema)
17. [API Reference](#17-api-reference)
18. [AI Agent System](#18-ai-agent-system)
19. [Chat System](#19-chat-system)
20. [Document Management](#20-document-management)
21. [Jobs System](#21-jobs-system)
22. [Developer Guide](#22-developer-guide)
23. [Known TODOs & Future Work](#23-known-todos--future-work)

---

## ⚠️ CRITICAL — Read Before Making Any Changes

> **This section is mandatory reading for every developer, AI agent, or LLM working on this codebase.**

### Multi-Developer & Multi-Agent Project

This project is actively developed by **multiple developers** and may be assisted by **multiple AI agents / LLMs** across different sessions. This documentation is the **single source of truth**. It is updated continuously as features are built and merged.

---

### Rules for AI Agents & LLMs

If you are an AI agent (Kiro, Copilot, Cursor, ChatGPT, Claude, Gemini, or any other) reading this document as context, you **MUST** follow these rules without exception:

#### 1. Never implement a feature that is not already in the codebase

If a feature is listed in [Section 23 — Known TODOs & Future Work](#23-known-todos--future-work), or is described as "planned", "placeholder", "coming soon", or "TODO" anywhere in this document, **do not create any code for it** unless the developer explicitly asks you to build it in the current session.

> **Why:** Another developer may already be working on it in a separate branch. Generating it independently will cause merge conflicts, duplicate logic, and inconsistent implementations.

#### 2. Check the codebase before creating anything new

Before creating a new component, store, API helper, or page — **search the existing codebase first**. If it already exists, extend or modify it. Do not create a duplicate.

#### 3. Never rename or restructure existing files without being asked

The folder structure, file names, and import paths documented here are intentional. Renaming or moving files without an explicit instruction will break imports across the project.

#### 4. Never add a new dependency without being asked

Do not add entries to `package.json` unless the developer explicitly requests a new library. Check if an existing dependency already covers the need.

#### 5. Match the existing code style exactly

- Use `cn()` from `@/lib/utils` for all class merging — never string concatenation
- Use `@/` path aliases — never relative `../../` imports
- All API calls go through `lib/auth.ts` wrappers — never call `api` directly from components
- All shared types go in `types/index.ts` — never inline `any`
- Add `"use client"` only when the component uses hooks, events, or browser APIs

#### 6. This document reflects what is BUILT, not what is planned

Every component, route, store, and API endpoint described in this README **exists in the codebase right now** (at the time of last update). Sections marked TODO or planned do **not** exist yet.

#### 7. When in doubt, ask — do not assume

If the developer's instruction is ambiguous, ask for clarification before writing code. A wrong assumption that gets committed is harder to fix than a 10-second question.

---

### For Human Developers

- **Before starting a new feature:** check Section 23 (TODOs) and confirm it is not already assigned to another developer.
- **After completing a feature:** update this README — remove it from Section 23 and add it to the relevant section with full documentation.
- **After adding a new UI component:** document it in [Section 8 — Component Library](#8-component-library) with props table and usage example.
- **After adding a new API endpoint:** document it in [Section 17 — API Reference](#17-api-reference).
- **After adding a new store or changing state shape:** update [Section 9 — State Management](#9-state-management).

> Keeping this document accurate is as important as writing the code itself. Outdated docs cause the same problems as no docs.

---

## 1. Project Overview

Visionary is an AI-powered learning platform. Users sign up, complete onboarding (selecting their role, grade, and board), then start learning subjects. An AI tutor answers questions via streaming chat, generates Mermaid diagrams, and suggests resources. Progress is tracked per chapter.

**User roles:** Student · Teacher · Organization

**Core flows:**
- Sign up / Login — email + OTP or Google OAuth
- Onboarding — category → grade / board / subject details
- Dashboard — subject cards → start a subject
- Learn — continue learning, chapter progress
- Chat — AI tutor with streaming, diagrams, resources (UI placeholder, backend complete)
- Profile — account settings, password change, notification preferences

---

## 2. Tech Stack

| Layer | Technology |
|---|---|
| Backend framework | NestJS 10 |
| ORM | Prisma 5 |
| Database | PostgreSQL (Supabase) |
| Cache / Queue | Redis (ioredis), BullMQ (planned) |
| Auth | Passport.js — JWT + Google OAuth 2.0 |
| AI | OpenAI SDK (GPT-4o, GPT-4o-mini) |
| File storage | Supabase Storage |
| Email | Nodemailer (SMTP) |
| Frontend framework | Next.js 16 (App Router) |
| UI | React 19, Tailwind CSS 4 |
| State | Zustand 5 (persisted) |
| Server state | TanStack React Query 5 |
| Charts | ApexCharts (react-apexcharts) |
| Icons | Lucide React, React Icons |
| HTTP client | Axios |
| Node version | 20.x |

---

## 3. Repository Structure

```
/
├── backend/
│   ├── prisma/
│   │   ├── schema.prisma          # Full DB schema
│   │   ├── seed.ts                # Seed master data
│   │   └── migrations/            # Prisma migration history
│   └── src/
│       ├── main.ts                # Bootstrap, CORS, global prefix /api
│       ├── app.module.ts          # Root module
│       ├── auth/                  # JWT + Google OAuth + OTP
│       ├── onboarding/            # Onboarding completion
│       ├── boards/                # BoardMaster CRUD
│       ├── grades/                # GradeMaster CRUD
│       ├── subjects/              # Subject listing, start, chapter tracking
│       ├── chat/                  # Conversations, messages, SSE streaming
│       ├── documents/             # PDF upload → Supabase Storage
│       ├── jobs/                  # Async job creation/status
│       ├── orchestrator/          # AI intent detection + agent routing
│       ├── agents/
│       │   ├── content/           # GPT-4o streaming explanations
│       │   ├── diagram/           # GPT-4o-mini Mermaid diagrams
│       │   └── resource/          # YouTube resource suggestions (placeholder)
│       ├── email/                 # Nodemailer OTP emails
│       ├── prisma/                # PrismaService (singleton)
│       └── redis/                 # RedisService (ioredis wrapper)
│
└── frontend/
    ├── app/
    │   ├── (auth)/                # Login, Signup, OTP pages — no layout
    │   ├── (dashboard)/           # Protected layout + Dashboard, Learn, Chat, Profile
    │   ├── auth/callback/         # Google OAuth callback handler
    │   └── onboarding/            # Category + Details pages
    ├── features/
    │   ├── dashboard/             # SubjectCard, ProgressCards, DashboardDetails
    │   ├── learn/                 # ContinueLearning, YourSubjects, detail view
    │   └── profile/               # ProfilePage shell + AccountTab + PreferencesTab
    ├── components/
    │   ├── shared/                # QueryProvider, OnboardingShell
    │   └── ui/                    # All reusable UI primitives
    ├── store/
    │   ├── authStore.ts           # Zustand auth store (persisted)
    │   ├── chatStore.ts           # Zustand chat store
    │   └── uiStore.ts             # Zustand UI store (sidebar state)
    ├── lib/
    │   ├── api.ts                 # Axios instance with auth interceptor
    │   ├── auth.ts                # All API call helpers
    │   └── utils.ts               # cn() utility (clsx + tailwind-merge)
    └── types/
        └── index.ts               # Shared TypeScript interfaces
```

---

## 4. Environment Setup

### Backend — `backend/.env`

```env
NODE_ENV=development
PORT=3001

# PostgreSQL (Supabase)
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.[PROJECT_REF].supabase.co:5432/postgres
DIRECT_URL=postgresql://postgres:[PASSWORD]@db.[PROJECT_REF].supabase.co:5432/postgres

# JWT
JWT_SECRET=your_long_random_secret
JWT_EXPIRES_IN=7d

# Redis
REDIS_URL=redis://localhost:6379

# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_CALLBACK_URL=http://localhost:3001/api/auth/google/callback

# OpenAI
OPENAI_API_KEY=sk-...

# Supabase Storage
SUPABASE_URL=https://[PROJECT_REF].supabase.co
SUPABASE_SERVICE_KEY=your_service_role_key
SUPABASE_STORAGE_BUCKET=uploads

# SMTP (for OTP emails)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_SECURE=false
SMTP_USER=user@example.com
SMTP_PASS=password
SMTP_FROM=Visionary <noreply@visionary.com>

# Frontend (CORS)
FRONTEND_URL=http://localhost:3000
```

### Frontend — `frontend/.env`

```env
NEXT_PUBLIC_API_URL=http://localhost:3001
```

---

## 5. Running the Project

### Prerequisites

- Node 20.x (`nvm use` or `.nvmrc`)
- PostgreSQL (Supabase recommended)
- Redis (local or Upstash)

### Backend

```bash
cd backend
npm install
cp .env.example .env        # fill in all values
npx prisma migrate dev       # run migrations
npx ts-node prisma/seed.ts   # seed boards, grades, subjects
npm run start:dev            # http://localhost:3001/api
```

### Frontend

```bash
cd frontend
npm install
# create .env with NEXT_PUBLIC_API_URL=http://localhost:3001
npm run dev                  # http://localhost:3000
```

### Useful Prisma Commands

```bash
npm run prisma:generate   # regenerate client after schema changes
npm run prisma:migrate    # run new migrations
npm run prisma:studio     # open Prisma Studio GUI
npm run prisma:seed       # re-seed master data
```

---

## 6. Frontend Architecture

### Overview

Next.js 16 App Router. Pages are React Server Components by default; interactive pages opt in with `"use client"`. The app is split into **route groups**, **feature modules**, and a **shared UI library**.

### Data Flow

```
User interaction
  → React component (feature)
      → lib/auth.ts API helper
          → lib/api.ts (Axios + auth interceptor)
              → NestJS backend /api/*
                  → Response mapped → Zustand store updated
                      → UI re-renders
```

### Route Groups

```
app/
├── (auth)/                    # Full-page, no shared layout
│   ├── login/page.tsx         # Login options (email / Google)
│   ├── login/email/page.tsx   # Email + password form
│   ├── signup/page.tsx        # Signup options
│   ├── signup/email/page.tsx  # Email + password registration
│   └── signup/otp/page.tsx    # 4-digit OTP verification
│
├── (dashboard)/               # Shared DashboardLayout (header + sidebar)
│   ├── layout.tsx             # Header, sidebar, mobile drawer
│   ├── dashboard/page.tsx     # Pre-start: subject cards | Post-start: DashboardDetails
│   ├── learn/page.tsx         # ContinueLearning + YourSubjects
│   ├── learn/[id]/page.tsx    # Subject detail + chapter list
│   ├── learn/continue-learning/page.tsx
│   ├── chat/[id]/page.tsx     # Chat UI (placeholder)
│   └── profile/page.tsx       # Account settings
│
├── auth/callback/page.tsx     # Google OAuth token handler
│
└── onboarding/
    ├── page.tsx               # Category selection
    └── details/page.tsx       # Role-specific details form
```

### Feature Modules

Each feature lives in `features/<name>/` and owns its components, sub-components, and local utilities. Pages in `app/` are thin wrappers that import from features.

| Feature | Path | Responsibility |
|---|---|---|
| Dashboard | `features/dashboard/` | SubjectCard, ProgressCards, DashboardDetails, StatCard, ContinueLearningCard, MyLessons |
| Learn | `features/learn/` | ContinueLearning, YourSubjects, detail view with Accordion chapters |
| Profile | `features/profile/` | ProfilePage shell, AccountTab, PreferencesTab, shared primitives |

### lib/api.ts

```typescript
// Axios instance — base URL from NEXT_PUBLIC_API_URL
// Request interceptor: reads token from localStorage "auth-storage" → sets Authorization header
// Response interceptor: on 401 → clears storage + cookie → redirects to /login
```

### lib/auth.ts

All typed API wrappers. Every function calls `api.ts` and returns typed data:

```typescript
loginWithEmail(email, password)       → { token, user }
signupWithEmail(email, password)      → { message, email }
verifyOtp(email, otp)                 → { token, user }
resendOtp(email)                      → { message }
completeOnboarding(payload)           → { token, user }
getOnboardingStatus()                 → User
getBoards()                           → BoardMaster[]
getGrades()                           → GradeMaster[]
getMySubjects()                       → SubjectMaster[]
startSubject(id)                      → { token, user }
getStartedSubject()                   → StartedSubject | null
getSubjectChapters(id)                → BookWithChapters[]
getAllStartedChapters()               → StartedSubjectTrack[]
getGoogleAuthUrl()                    → string
```

---

## 7. Design System

The design system is documented in full at `frontend/DESIGN_SYSTEM.md`. Key tokens are summarised here.

### Color Palette

| Role | Class / Value | Hex |
|---|---|---|
| Primary | `bg-blue-600` / `text-blue-600` | `#2563EB` |
| Primary hover | `hover:bg-blue-700` | `#1D4ED8` |
| Light blue bg | `bg-blue-50` | `#EFF6FF` |
| Page background | `bg-white` | `#FFFFFF` |
| Dashboard bg | `bg-gray-50` | `#F9FAFB` |
| Profile bg | — | `#F0F0F0` |
| Border default | `border-gray-200` | `#E5E7EB` |
| Text primary | `text-gray-900` | `#111827` |
| Text secondary | `text-gray-600` | `#4B5563` |
| Text muted | `text-gray-400` | `#9CA3AF` |
| Error | `text-red-500` | `#EF4444` |
| Disabled | `bg-gray-300` | `#D1D5DB` |

CSS custom properties (`app/globals.css`):

```css
--color-primary:       #2563EB;
--color-primary-hover: #1D4ED8;
```

### Typography

Font: **Inter** (Google Fonts, loaded in `app/layout.tsx`).

| Usage | Classes |
|---|---|
| Page heading | `text-xl sm:text-2xl font-bold text-gray-900` |
| Section heading | `text-lg font-semibold text-gray-900` |
| Body | `text-sm text-gray-700` |
| Caption / muted | `text-xs text-gray-400` |
| Button label | `text-sm font-medium` |

### Spacing & Sizing

| Element | Value |
|---|---|
| Auth card max-width | `988px` |
| Auth card border-radius | `42px` |
| Dashboard header height | `h-16` (64px) |
| Desktop sidebar width | `w-16` collapsed → `w-52` expanded |
| Pill button height | `56px`, `border-radius: 50px` |
| Input padding | `21px 24px` |
| Profile tab card | `256×192px`, `border-radius: 24px`, `padding: 32px` |

### Border Radius

| Element | Value |
|---|---|
| Auth / onboarding cards | `42px` |
| Pill buttons | `50px` |
| Standard inputs | `rounded-lg` (8px) |
| Dashboard / profile cards | `rounded-2xl` (16px) |
| Avatar | `rounded-full` |

### Elevation

No box-shadows. Elevation is communicated through borders only:

- Auth cards: `border: 1px solid rgba(0,0,0,0.12)`
- Dashboard cards: `border border-gray-200`
- Inputs default: `border border-gray-300`
- Inputs focused: `focus:border-blue-600`
- Inputs error: `border-red-400`

---
The design system is documented in full at `frontend/DESIGN_SYSTEM.md`. Key tokens are summarised here.

### Color Palette

| Role | Class / Value | Hex |
|---|---|---|
| Primary | `bg-blue-600` / `text-blue-600` | `#2563EB` |
| Primary hover | `hover:bg-blue-700` | `#1D4ED8` |
| Light blue bg | `bg-blue-50` | `#EFF6FF` |
| Page background | `bg-white` | `#FFFFFF` |
| Dashboard bg | `bg-gray-50` | `#F9FAFB` |
| Profile bg | — | `#F0F0F0` |
| Border default | `border-gray-200` | `#E5E7EB` |
| Text primary | `text-gray-900` | `#111827` |
| Text secondary | `text-gray-600` | `#4B5563` |
| Text muted | `text-gray-400` | `#9CA3AF` |
| Error | `text-red-500` | `#EF4444` |
| Disabled | `bg-gray-300` | `#D1D5DB` |

CSS custom properties (defined in `app/globals.css`):

```css
--color-primary:       #2563EB;
--color-primary-hover: #1D4ED8;
```

### Typography

Font: **Inter** (Google Fonts, loaded in `app/layout.tsx`). Tailwind utility classes are used directly — no custom font tokens beyond the global body font.

| Usage | Classes |
|---|---|
| Page heading | `text-xl sm:text-2xl font-bold text-gray-900` |
| Section heading | `text-lg font-semibold text-gray-900` |
| Body | `text-sm text-gray-700` |
| Caption / muted | `text-xs text-gray-400` |
| Button label | `text-sm font-medium` |

### Spacing & Sizing

| Element | Value |
|---|---|
| Auth card max-width | `988px` |
| Auth card border-radius | `42px` |
| Onboarding card max-width | `800px` |
| Dashboard header height | `h-16` (64px) |
| Desktop sidebar width | `w-16` → `w-52` (expanded) |
| Button height (pill) | `56px`, `border-radius: 50px` |
| Input padding | `21px 24px` |
| Profile tab card | `256×192px`, `border-radius: 24px`, `padding: 32px` |

### Border Radius Tokens

| Element | Value |
|---|---|
| Cards / modals | `42px` |
| Pill buttons | `50px` |
| Standard inputs | `rounded-lg` (8px) |
| Dashboard cards | `rounded-2xl` (16px) |
| Notification cards | `rounded-2xl` (16px) |
| Avatar | `rounded-full` |

### Elevation

No box-shadows. Elevation is communicated through borders only:

- Auth cards: `border: 1px solid rgba(0,0,0,0.12)`
- Dashboard cards: `border border-gray-200`
- Inputs default: `border border-gray-300`
- Inputs focused: `focus:border-blue-600`
- Inputs error: `border-red-400`

---

## 8. Component Library

All reusable primitives live in `frontend/components/ui/`. Import path: `@/components/ui/<Name>`.

---

### Accordion

**Purpose:** Collapsible content sections with animated open/close.

| Prop | Type | Default | Description |
|---|---|---|---|
| `items` | `AccordionItem[]` | required | `{ id, title, content, onClick? }[]` |
| `defaultOpen` | `string[]` | `[]` | IDs open on mount |
| `allowMultiple` | `boolean` | `false` | Allow multiple open simultaneously |
| `activeId` | `string?` | — | Highlights an item as active (bold) |

```tsx
<Accordion
  items={[{ id: "1", title: "Chapter 1", content: <p>Content</p> }]}
  defaultOpen={["1"]}
/>
```

---

### Badge

**Purpose:** Small label pill with optional icon or image.

| Prop | Type | Default | Description |
|---|---|---|---|
| `label` | `string` | required | Text content |
| `imageSrc` | `string?` | — | 16×16 image icon |
| `icon` | `ReactNode?` | — | JSX icon (used if no imageSrc) |
| `variant` | `"outline" \| "gray"` | `"outline"` | Visual style |
| `className` | `string?` | — | Extra classes |

```tsx
<Badge label="Science" variant="gray" />
<Badge label="CBSE" imageSrc="/icons/cbse.png" />
```

---

### BarChart

**Purpose:** ApexCharts bar chart, SSR-safe via dynamic import.

| Prop | Type | Default | Description |
|---|---|---|---|
| `series` | `BarChartSeries[]` | required | `{ name, data: [{x,y}][], color? }[]` |
| `height` | `number` | `240` | Chart height in px |
| `primaryColor` | `string` | `#2563EB` | First series color |
| `secondaryColor` | `string` | `#93C5FD` | Second series color |

---

### Breadcrumb

**Purpose:** Navigation trail — home icon → chevron → page labels.

| Prop | Type | Default | Description |
|---|---|---|---|
| `items` | `BreadcrumbItem[]` | required | `{ label, href? }[]` — last item is current page |
| `className` | `string?` | — | Extra classes on `<nav>` |

- Home icon always navigates to `/dashboard`
- Last item renders as `<span>`, intermediate items as `<Link>`

```tsx
<Breadcrumb
  items={[{ label: "Learn", href: "/learn" }, { label: "Account settings" }]}
  className="mb-6"
/>
```

---

### Dialog

**Purpose:** Full-screen modal (92vw × 92vh) with header, scrollable body, and bottom input bar.

| Prop | Type | Description |
|---|---|---|
| `open` | `boolean` | Controls visibility |
| `onClose` | `() => void` | Called on Escape or minimize click |
| `title` | `string?` | Header left label |
| `children` | `ReactNode` | Scrollable body content |

```tsx
<Dialog open={isOpen} onClose={() => setIsOpen(false)} title="Chapter 6.1">
  <p>Chapter content here...</p>
</Dialog>
```

---

### DonutChart

**Purpose:** ApexCharts donut chart with center label, SSR-safe.

| Prop | Type | Default | Description |
|---|---|---|---|
| `series` | `number[]` | `[35.1, 23.5, 2.4, 5.4]` | Slice values |
| `labels` | `string[]` | — | Slice labels |
| `height` | `number` | `280` | Chart height |
| `centerLabel` | `string` | `"Overall Progress"` | Text in donut center |
| `colors` | `string[]` | blue palette | Slice colors |

---

### FloatingInput

**Purpose:** Animated floating-label input. Label transitions from center to top-left on focus or when a value is present.

| Prop | Type | Description |
|---|---|---|
| `label` | `string` | Floating label text |
| `error` | `string?` | Error message (turns border + label red) |
| All native `<input>` props | — | Forwarded via `forwardRef` |

- Password type: shows eye/eye-off toggle automatically
- Error state: red border + red label + info icon message below

```tsx
<FloatingInput
  label="Email address"
  type="email"
  value={email}
  onChange={(e) => setEmail(e.target.value)}
  error={errors.email}
/>
```

---

### Popover

**Purpose:** Floating panel anchored to a trigger element. Supports controlled and uncontrolled modes. Closes on outside click and Escape.

| Prop | Type | Default | Description |
|---|---|---|---|
| `trigger` | `ReactNode` | required | Element that opens the popover |
| `children` | `ReactNode` | required | Panel content |
| `align` | `"left" \| "right" \| "center"` | `"right"` | Horizontal alignment |
| `side` | `"top" \| "bottom"` | `"bottom"` | Which side to open on |
| `className` | `string?` | — | Extra classes on panel |
| `open` | `boolean?` | — | Controlled open state |
| `onOpenChange` | `(open: boolean) => void` | — | Controlled change handler |

Sub-components: `PopoverHeader`, `PopoverBody`, `PopoverItem`

```tsx
<Popover
  trigger={<button>Open</button>}
  align="right"
>
  <PopoverHeader title="Menu" />
  <PopoverItem icon={<User size={15} />} onClick={handleProfile}>Profile</PopoverItem>
  <PopoverItem icon={<LogOut size={15} />} variant="danger" onClick={handleLogout}>Logout</PopoverItem>
</Popover>
```

---

### ProgressBar

**Purpose:** Simple horizontal progress bar.

| Prop | Type | Description |
|---|---|---|
| `percentage` | `number` | 0–100 fill value |

```tsx
<ProgressBar percentage={72} />
```

---

### ScrollArea

**Purpose:** Custom scrollable container with a styled scrollbar thumb.

| Prop | Type | Description |
|---|---|---|
| `children` | `ReactNode` | Scrollable content |
| `className` | `string?` | Extra classes on wrapper |

---

### Tabs

**Purpose:** Tab switcher with two visual variants.

| Prop | Type | Default | Description |
|---|---|---|---|
| `tabs` | `string[]` | required | Tab labels |
| `activeTab` | `string` | required | Currently active label |
| `onChange` | `(tab: string) => void` | required | Change handler |
| `variant` | `"underline" \| "pill"` | `"underline"` | Visual style |

- `"underline"` — bottom border indicator, used in content areas
- `"pill"` — `190×56px` rounded pill buttons, used in the Profile sidebar

```tsx
// Underline (default)
<Tabs tabs={["Overview", "Details"]} activeTab={active} onChange={setActive} />

// Pill (profile sidebar)
<Tabs variant="pill" tabs={["Account", "Preferences"]} activeTab={active} onChange={setActive} />
```

---

### Toggle

**Purpose:** Accessible checkbox-based toggle switch.

| Prop | Type | Description |
|---|---|---|
| `checked` | `boolean` | Current state |
| `onChange` | `(checked: boolean) => void` | Change handler |
| `label` | `string?` | Optional text label to the right |
| `className` | `string?` | Extra classes on wrapper |

- White track with blue border when on, gray border when off
- Solid blue thumb when on, gray when off
- Uses hidden `<input type="checkbox">` + `useId()` for accessibility

```tsx
<Toggle checked={enabled} onChange={setEnabled} label="Daily reminders" />
```

---

### Tooltip

**Purpose:** Hover tooltip with directional arrow.

| Prop | Type | Default | Description |
|---|---|---|---|
| `content` | `string` | required | Tooltip text |
| `children` | `ReactNode` | required | Trigger element |
| `position` | `"top" \| "bottom" \| "left" \| "right"` | `"top"` | Tooltip placement |

```tsx
<Tooltip content="Coming soon" position="bottom">
  <button disabled>Practice</button>
</Tooltip>
```

---

## 9. State Management

The app uses three Zustand stores. All stores are in `frontend/store/`.

---

### authStore (`store/authStore.ts`)

Persisted to `localStorage` under key `"auth-storage"`. On rehydration, re-sets the `auth_token` cookie so the Next.js middleware can read it server-side.

```typescript
interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  setAuth(token: string, user: User): void  // saves token, sets cookie, updates state
  clear(): void                              // deletes cookie, clears state → triggers redirect
}
```

Cookie: `auth_token`, 7-day expiry, `SameSite=Lax`, path `/`.

---

### chatStore (`store/chatStore.ts`)

In-memory only (not persisted). Manages streaming chat state.

```typescript
interface ChatState {
  chats: Chat[]
  activeChatId: string | null
  streaming: boolean
  setChats(chats: Chat[]): void
  setActiveChat(id: string): void
  addMessage(chatId: string, message: ChatMessage): void
  appendToLastMessage(chatId: string, chunk: string): void  // used during SSE streaming
  setStreaming(value: boolean): void
}
```

---

### uiStore (`store/uiStore.ts`)

In-memory only. Controls sidebar open/close state.

```typescript
interface UIState {
  sidebarOpen: boolean
  setSidebarOpen(open: boolean): void
  toggleSidebar(): void
}
```

---

### User Type

```typescript
interface User {
  id: string
  email: string
  fullName?: string
  category?: "student" | "teacher" | "organization"
  grade?: string
  board?: string
  subject?: string
  organizationName?: string
  organizationType?: string
  onboardingCompleted: boolean
  isStarted?: boolean
}
```

---

## 10. Routing & Middleware

### Middleware (`middleware.ts`)

Runs on every matched route. Reads the `auth_token` cookie, decodes the JWT payload (no verification — edge runtime), and applies redirect rules:

```
No token
  → allow auth pages (/login, /signup)
  → redirect everything else → /login

Valid token, onboardingCompleted = false
  → redirect protected pages + auth pages → /onboarding

Valid token, onboardingCompleted = true
  → redirect /onboarding → /dashboard
  → redirect /login, /signup → /dashboard
  → allow all other routes

Special cases (always allowed):
  → /auth/callback  (Google OAuth return)
  → /signup/otp     (OTP entry page)
```

Matcher covers: `/login`, `/signup`, `/onboarding`, `/dashboard`, `/chat` and all sub-paths.

### Dashboard Conditional Rendering

```
/dashboard
  user.isStarted === false  →  Subject cards + ProgressCards (start learning)
  user.isStarted === true   →  DashboardDetails (stats, continue learning, lessons)
```

### Navigation Items

| Label | Path | Status |
|---|---|---|
| Dashboard | `/dashboard` | Active |
| Learn | `/learn` | Active |
| Ask | `/chat` | Disabled — coming soon |
| Practice | `/practice` | Disabled — coming soon |

---

## 11. Auth Flow

### Email + Password Registration

```
User fills email + password
  → POST /api/auth/register
  → Backend: hash password, generate 4-digit OTP, store with 10-min expiry
  → Backend: send OTP email via Nodemailer
  → Frontend: redirect to /signup/otp?email=...
  → User enters OTP (4 boxes, auto-advance, paste support, 60s resend timer)
  → POST /api/auth/verify-otp
  → Backend: validate OTP + expiry, mark isVerified=true, clear OTP fields
  → Backend: issue JWT
  → Frontend: store token+user in Zustand, redirect → middleware → /onboarding or /dashboard
```

### Email + Password Login

```
User fills email + password
  → POST /api/auth/login
  → Backend: bcrypt.compare, check isVerified, issue JWT
  → Frontend: store token+user, redirect based on onboardingCompleted
```

### Google OAuth

```
User clicks Google button
  → GET /api/auth/google (Passport redirects to Google)
  → Google redirects to /api/auth/google/callback
  → Backend: upsert user (googleId), issue JWT
  → Backend: redirect to frontend /auth/callback?token=...&onboardingCompleted=...
  → Frontend CallbackHandler: decode JWT payload, store in Zustand
  → Redirect to /onboarding or /dashboard
```

### JWT Structure

```json
{
  "sub": "<userId>",
  "email": "user@example.com",
  "onboardingCompleted": true,
  "isStarted": false,
  "iat": 1234567890,
  "exp": 1235172690
}
```

Token is stored in:
- Zustand store (in-memory + localStorage via `persist`)
- `auth_token` cookie (7-day, SameSite=Lax) — used by middleware for SSR route protection

---

## 12. Onboarding Flow

```
/onboarding (category selection)
  → User picks: Student | Teacher | Organization
  → Stored in sessionStorage as "onboarding_category"
  → Navigate to /onboarding/details

/onboarding/details
  → Loads boards + grades from API
  → Student: fullName + grade (dropdown) + board (CBSE / ICSE / Other)
  → Teacher: fullName + subject
  → Organization: fullName + orgName + orgType
  → POST /api/onboarding/complete
  → Backend: update user fields, re-issue JWT with onboardingCompleted: true
  → Frontend: setAuth with new token, redirect to /dashboard
```

---

## 13. Learning System

### Data Hierarchy

```
GradeMaster (1st–12th)
  └── SubjectMaster (Math, Science, etc.)
       ├── boardId: null  → common for all boards
       └── boardId: <id>  → board-specific
            └── BookMaster (e.g. "NCERT Part 1")
                 └── ChapterMaster (e.g. "Chapter 1: Real Numbers")
```

### Subject Resolution

When fetching subjects for a user:
1. Try board-specific subjects (`gradeId = user.grade AND boardId = user.board`)
2. If none found, fall back to common subjects (`gradeId = user.grade AND boardId = null`)

### Starting a Subject

```
User clicks "Start" on a subject card
  → POST /api/subjects/:id/start
  → Backend: upsert SubjectTrack, set user.isStarted = true, re-issue JWT
  → Frontend: update auth store with new JWT
  → Dashboard re-renders to DashboardDetails view
```

### Chapter Progress

ChapterTrack stores:
- `progress`: 0–100 integer
- `status`: `NOT_STARTED | IN_PROGRESS | COMPLETED`
- `startedAt`, `completedAt`

Progress is merged into chapter lists on every `getChapters` / `getAllStartedChapters` call.

---

## 14. Profile & Settings

Route: `/profile` (inside dashboard layout).

### Structure

```
features/profile/
├── index.tsx                  # Shell: breadcrumb + tab sidebar + renders tabs
└── components/
    ├── shared.tsx             # SectionCard, PillButton, Toggle (re-exported)
    ├── AccountTab.tsx         # Basic Details + Password + Delete Account
    └── PreferencesTab.tsx     # Appearance + Email notifications
```

### AccountTab

Three sections:

1. **Basic Details** — 2-column grid: Full name, Email (read-only), School name, Board, Grade, Phone (with `+91` dial-code prefix). "Update" pill button (disabled, ready to wire).

2. **Password** — 3-column row: Current / New / Retype password with FloatingInput password toggle. "Change password" pill button with match validation.

3. **Delete Account** — warning copy + red "Delete your account" pill button.

### PreferencesTab

Four sections:

1. **Appearance** — segmented control: Dark / Light / Auto (white active pill on gray track).

2. **Email notifications** heading (outside any card).

3. **Streaks** card — "Streaks daily reminder" toggle.

4. **Learning reminders** card — "Daily Practice" toggle.

5. **News & Announcements** card — "Product launches and updates" + "Offers & Promotions" toggles.

All form state is local and ready to wire to API endpoints once the backend exposes them.

---

## 15. Backend Architecture

The backend is a NestJS monolith running on port `3001` with global prefix `/api`.

### Module Map

```
AppModule
├── ConfigModule (global)
├── PrismaModule (global singleton)
├── RedisModule
├── AuthModule
│   └── Strategies: JwtStrategy, GoogleStrategy
├── OnboardingModule
├── BoardsModule
├── GradesModule
├── SubjectsModule
├── ChatModule
│   └── OrchestratorModule
│       ├── ContentAgentModule
│       ├── DiagramAgentModule
│       └── ResourceAgentModule
├── DocumentsModule
├── EmailModule
└── JobsModule
```

### Key Services

**PrismaService** — extends `PrismaClient`, connects on `onModuleInit`, disconnects on `onModuleDestroy`.

**RedisService** — ioredis wrapper with `lazyConnect: true` (won't crash if Redis is offline). Exposes `get`, `set`, `del`, `incr`, `expire`.

**AuthService** — `register` (hash password, generate OTP, send email), `verifyOtp`, `resendOtp`, `login` (bcrypt compare), `googleLogin` (upsert user), `issueToken` (sign JWT with `sub`, `email`, `onboardingCompleted`, `isStarted`).

**OnboardingService** — updates user fields, re-issues JWT with `onboardingCompleted: true`.

**SubjectsService** — `getForUser` returns board-specific subjects first, falls back to common (`boardId=null`). `startSubject` upserts SubjectTrack, sets `isStarted=true`, re-issues JWT. `getChapters` merges ChapterTrack progress into chapter list.

**OrchestratorService** — classifies prompt intent via GPT-4o-mini (JSON: `{explanation, diagram, image, resource}`), then always streams explanation via ContentAgent, conditionally fires DiagramAgent and ResourceAgent in parallel via `Promise.allSettled`.

**ChatService** — saves user message, streams via orchestrator, saves assistant message. Uses RxJS Observable for SSE.

**DocumentsService** — uploads to Supabase Storage, creates Document record with status `PROCESSING`.

**EmailService** — Nodemailer transporter, sends branded OTP email with 10-minute expiry notice.

---

## 16. Database Schema

All tables use PostgreSQL via Supabase. Prisma manages migrations.

### Users

| Field | Type | Notes |
|---|---|---|
| id | cuid | Primary key |
| email | String (unique) | |
| name | String | Display name |
| avatar | String? | URL |
| passwordHash | String? | Null for Google users |
| googleId | String? (unique) | Google OAuth |
| plan | Enum FREE/PRO | Default FREE |
| isVerified | Boolean | Email verified via OTP |
| otp | String? | 4-digit code |
| otpExpiry | DateTime? | 10-minute window |
| onboardingCompleted | Boolean | |
| isStarted | Boolean | True after first subject started |
| category | String? | student / teacher / organization |
| fullName | String? | From onboarding |
| grade | String? | GradeMaster.id |
| board | String? | BoardMaster.id |
| subject | String? | Teacher's subject |
| organizationName | String? | |
| organizationType | String? | |

### Master Data (seeded)

- **BoardMaster** — CBSE, ICSE, and other boards
- **GradeMaster** — 1st through 12th grade
- **SubjectMaster** — linked to grade + optional board (`null` = common for all boards), has icon
- **BookMaster** — linked to subject, ordered by `sortOrder`
- **ChapterMaster** — linked to book, ordered by `sortOrder`

### User Progress

- **SubjectTrack** — records when a user starts a subject (`userId + subjectId` unique)
- **ChapterTrack** — per-chapter progress (0–100), status: `NOT_STARTED | IN_PROGRESS | COMPLETED`

### Chat

- **Conversation** — belongs to user, has title
- **Message** — belongs to conversation, role: `user | assistant`, stores content, diagram (Mermaid), resources (JSON), imageUrl

### Documents

- **Document** — PDF upload record, status: `PROCESSING | READY | ERROR`
- **DocumentChunk** — text chunks for RAG; embedding column added manually via pgvector SQL

### Jobs

- **Job** — async task record, type: `IMAGE_GENERATION | PDF_PROCESSING`, status: `PENDING | COMPLETED | FAILED`

---

## 17. API Reference

All endpoints are prefixed with `/api`. JWT-protected routes require `Authorization: Bearer <token>`.

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | No | Register with email+password. Sends OTP. Returns `{message, email}` |
| POST | `/auth/verify-otp` | No | Verify 4-digit OTP. Returns `{accessToken, user}` |
| POST | `/auth/resend-otp` | No | Resend OTP to email |
| POST | `/auth/login` | No | Login with email+password. Returns `{accessToken, user}` |
| GET | `/auth/google` | No | Redirect to Google OAuth consent screen |
| GET | `/auth/google/callback` | No | Google OAuth callback → redirects to `/auth/callback?token=...` |
| GET | `/auth/me` | JWT | Returns current user from JWT payload |

### Onboarding

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/onboarding/status` | JWT | Returns user's onboarding fields |
| POST | `/onboarding/complete` | JWT | Save onboarding data, returns new `{accessToken, user}` |

```json
// POST /onboarding/complete body
{
  "category": "student",
  "fullName": "Jane Doe",
  "grade": "<GradeMaster.id>",
  "board": "<BoardMaster.id>"
}
```

### Master Data

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/boards` | No | List all active boards |
| GET | `/grades` | No | List all active grades |

### Subjects & Learning

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/subjects/me` | JWT | Subjects for user's grade+board |
| GET | `/subjects/started` | JWT | Most recently started subject |
| GET | `/subjects/my-tracks` | JWT | All started subjects with books+chapters+progress |
| POST | `/subjects/:id/start` | JWT | Start a subject, returns new JWT with `isStarted: true` |
| GET | `/subjects/:id/chapters` | JWT | Books+chapters+progress for a subject |

### Chat

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/chat/conversations` | JWT | Create new conversation |
| GET | `/chat/conversations` | JWT | List user's conversations (last 50) |
| GET | `/chat/conversations/:id/messages` | JWT | Get messages for a conversation |
| GET (SSE) | `/chat/stream?prompt=...&conversationId=...` | JWT | Stream AI response tokens |

SSE event format:
```json
{ "token": "..." }    // streaming token
{ "done": true }      // stream complete
{ "error": "..." }    // error
```

Special inline tokens in the stream:
- `` [DIAGRAM]```mermaid\n...[/DIAGRAM] `` — Mermaid diagram block
- `[RESOURCES][{"title":"...","url":"..."}][/RESOURCES]` — resource links

### Documents

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/documents/upload` | JWT | Upload PDF (`multipart/form-data`, field: `file`) |
| GET | `/documents` | JWT | List user's documents |
| GET | `/documents/:id` | JWT | Get single document |

### Jobs

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/jobs/image` | JWT | Create image generation job |
| GET | `/jobs/:id` | JWT | Get job status |

---

## 18. AI Agent System

### Architecture

```
User prompt
  → OrchestratorService.streamResponse()
      → detectIntent() via GPT-4o-mini
          Returns: { explanation, diagram, image, resource }
      → Always: ContentAgentService.stream()   — GPT-4o streaming explanation
      → If diagram: DiagramAgentService.generate() — GPT-4o-mini Mermaid code
      → If resource: ResourceAgentService.fetch()  — placeholder YouTube links
      → All run via Promise.allSettled (parallel, non-blocking)
```

### ContentAgentService

- Model: `gpt-4o`
- System prompt: "You are an expert AI tutor. Explain concepts clearly with examples. Use markdown formatting."
- Streams tokens via callback

### DiagramAgentService

- Model: `gpt-4o-mini`
- Returns Mermaid code wrapped in `[DIAGRAM]...[/DIAGRAM]`
- Frontend parses and renders with a Mermaid renderer

### ResourceAgentService

- Currently a placeholder returning a static YouTube link
- TODO: Integrate YouTube Data API v3

---

## 19. Chat System

### Backend

- **Conversations** — created per session, stored with `userId`
- **Messages** — saved before (user) and after (assistant) streaming
- **SSE endpoint** — `GET /api/chat/stream?prompt=...&conversationId=...` returns RxJS Observable as SSE

### Frontend

The chat UI at `/chat/[id]` is currently a placeholder. The backend is fully implemented and ready to connect. The `chatStore` already handles streaming via `appendToLastMessage`.

---

## 20. Document Management

```
User uploads PDF
  → POST /documents/upload (multipart, field: "file")
  → Backend: upload to Supabase Storage at path userId/timestamp-filename.pdf
  → Create Document record with status PROCESSING
  → TODO: Queue PDF processing job (Phase 5)

RAG pipeline (planned):
  → Parse PDF text (pdf-parse)
  → Chunk text
  → Generate embeddings (OpenAI text-embedding-ada-002)
  → Store in document_chunks with pgvector embedding column
  → On chat query: cosine similarity search → inject context into prompt
```

pgvector setup (run manually in Supabase SQL editor after migration):

```sql
ALTER TABLE document_chunks ADD COLUMN embedding vector(1536);
CREATE INDEX ON document_chunks USING ivfflat (embedding vector_cosine_ops);
```

---

## 21. Jobs System

Async job tracking for long-running tasks.

| Job Type | Status |
|---|---|
| `IMAGE_GENERATION` | Pending — BullMQ queue integration planned (Phase 6) |
| `PDF_PROCESSING` | Pending — triggered after document upload |

Jobs are created with status `PENDING` and polled via `GET /jobs/:id`.

---

## 22. Developer Guide

### Adding a New Page

1. Create the route file inside the appropriate group:
   - Protected page → `app/(dashboard)/your-page/page.tsx`
   - Auth page → `app/(auth)/your-page/page.tsx`
2. If the page has significant logic, create a feature module: `features/your-feature/index.tsx`
3. The page file should be a thin wrapper:

```tsx
// app/(dashboard)/your-page/page.tsx
import YourFeature from "@/features/your-feature";
export default function YourPage() { return <YourFeature />; }
```

4. Add the route to the middleware matcher in `middleware.ts` if it needs protection.

---

### Adding a New UI Component

1. Create `frontend/components/ui/YourComponent.tsx`
2. Add `"use client"` if it uses hooks or browser APIs
3. Accept a `className` prop and merge with `cn()` for composability
4. Export as default

```tsx
"use client";
import { cn } from "@/lib/utils";

interface YourComponentProps {
  className?: string;
  // ...
}

export default function YourComponent({ className }: YourComponentProps) {
  return <div className={cn("base-classes", className)} />;
}
```

---

### Connecting to a New API Endpoint

1. Add a typed wrapper in `frontend/lib/auth.ts`:

```typescript
export async function getYourData(): Promise<YourType> {
  const { data } = await api.get<YourType>("/your-endpoint");
  return data;
}
```

2. Add the response type to `frontend/types/index.ts` if it's new.

3. Call it from your feature component:

```tsx
const [data, setData] = useState<YourType | null>(null);

useEffect(() => {
  getYourData().then(setData).catch(console.error);
}, []);
```

4. If the data is shared across components, add it to a Zustand store instead of local state.

---

### Adding a New Backend Module

1. Generate with NestJS CLI:

```bash
cd backend
nest g module your-module
nest g controller your-module
nest g service your-module
```

2. Import `PrismaModule` if you need DB access (it's global, so just inject `PrismaService`).
3. Add the module to `AppModule` imports.
4. Protect routes with `@UseGuards(JwtAuthGuard)`.

---

### Coding Standards

| Rule | Detail |
|---|---|
| Client components | Add `"use client"` only when needed (hooks, events, browser APIs) |
| Class merging | Always use `cn()` from `@/lib/utils` — never concatenate class strings manually |
| API calls | All calls go through `lib/auth.ts` wrappers — never call `api` directly from components |
| Types | All shared types live in `types/index.ts` — no inline `any` |
| Error handling | Wrap API calls in try/catch, show user-facing error messages |
| Form validation | Validate client-side before submitting, show field-level errors via `FloatingInput error` prop |
| Imports | Use `@/` path alias — never relative `../../` imports |

---

## 23. Known TODOs & Future Work

| Area | TODO |
|---|---|
| Chat UI | Build full chat interface at `/chat/[id]` — backend is ready |
| Profile | Wire AccountTab "Update" to `PATCH /auth/me` endpoint |
| Profile | Wire "Change password" to `PATCH /auth/password` endpoint |
| Profile | Implement avatar upload |
| Profile | Wire Preferences (appearance, notifications) to backend |
| Documents | PDF processing pipeline — parse, chunk, embed, store (Phase 5) |
| Documents | pgvector RAG — cosine similarity search on chat queries |
| Resources | YouTube Data API v3 integration in ResourceAgentService |
| Jobs | BullMQ queue integration for image generation (Phase 6) |
| Practice | Practice module (route disabled) |
| Ask | Ask module (route disabled) |
| Chapter progress | API endpoints to update ChapterTrack progress/status |
| Image generation | DALL-E or similar integration for `IMAGE_GENERATION` job type |
| Rate limiting | Redis-based prompt rate limiting (`promptsToday`, `lastResetAt` fields exist on User) |
| Plan gating | FREE vs PRO plan enforcement |
| Apple / Microsoft OAuth | Login buttons exist but are disabled |
| Dark mode | Appearance toggle in Preferences is UI-only — no dark mode CSS yet |

---

*Last updated: April 2026*
