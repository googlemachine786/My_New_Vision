# Visionary — Frontend Design System

> Stack: Svelte 5 · Tailwind CSS v4 · Vite · Product Sans

---

## 1. Typography

**Font family:** `Product Sans` (self-hosted TTF), aliased as `Google Sans` for fallback.  
Applied globally via `html` and `body` in `app.css`.

| Token class | Size | Weight | Line height |
|---|---|---|---|
| `.ds-h1` | 3.5rem (56px) | 700 | 1.1 |
| `.ds-h2` | 2.25rem (36px) | 700 | 1.2 |
| `.ds-h3` | 1.5rem (24px) | 700 | 1.3 |
| `.ds-body-lg` | 1.125rem (18px) | 400 | 1.6 |
| `.ds-body` | 1rem (16px) | 400 | 1.5 |
| `.ds-body-sm` | 0.875rem (14px) | 400 | 1.5 |
| `.ds-caption` | 0.75rem (12px) | 400 | 1.4 |

**Available weights:** 100 (Thin) · 300 (Light) · 400 (Regular) · 500 (Medium) · 700 (Bold) · 900 (Black) — each with italic variant.

**In-component usage pattern:**
```
font-medium text-4xl leading-[130%] tracking-[-0.005em]   ← page headings
font-medium text-base                                       ← buttons, links
text-sm text-gray-600                                       ← labels, captions
text-xs font-medium                                         ← sidebar nav labels
```

---

## 2. Color Palette

### Tailwind `@theme` tokens (used in CSS)

| Token | Value | Usage |
|---|---|---|
| `--color-primary` | `#2563EB` | Primary actions, links, active states |
| `--color-primary-hover` | `#1D4ED8` | Hover on primary |
| `--color-border` | `#D1D5DB` | Default input/card borders |
| `--color-text-primary` | `#111827` | Body text |
| `--color-text-secondary` | `#6B7280` | Muted text, placeholders |

### Tailwind utility classes used throughout

| Role | Class | Hex |
|---|---|---|
| Primary blue | `bg-[#2563EB]` / `text-blue-600` | `#2563EB` |
| Primary hover | `hover:bg-[#1D4ED8]` | `#1D4ED8` |
| Light blue bg | `bg-blue-50` / `#EFF6FF` | `#EFF6FF` |
| Blue border (selected) | `#BFDBFE` | `#BFDBFE` |
| Page background | `bg-white` | `#FFFFFF` |
| Dashboard bg | `bg-gray-50` | `#F9FAFB` |
| Border default | `border-gray-200` / `border-gray-300` | `#E5E7EB` / `#D1D5DB` |
| Border strong | `#8E8E93` | `#8E8E93` |
| Text primary | `text-gray-900` | `#111827` |
| Text secondary | `text-gray-600` | `#4B5563` |
| Text muted | `text-gray-400` / `text-gray-500` | `#9CA3AF` / `#6B7280` |
| Error | `text-red-500` / `border-red-400` | `#EF4444` / `#F87171` |
| Disabled | `bg-gray-300` / `#d1d5db` | `#D1D5DB` |
| Avatar blue | `bg-blue-500` | `#3B82F6` |

### CSS Variables (shadcn-style, HSL)

```css
--primary:            217 79% 44%   /* #2563EB */
--destructive:        5 70% 55%
--border:             0 0% 89.8%
--radius:             0.5rem
--sidebar-background: 0 0% 98%
--sidebar-border:     220 13% 91%
```

Dark mode overrides exist under `.dark {}` — all tokens shift to dark-surface values.

---

## 3. Spacing & Sizing

The app uses Tailwind's default spacing scale plus a set of recurring custom values applied via inline `style=`.

| Purpose | Value |
|---|---|
| Card padding (desktop) | `72px 42px 48px` |
| Card padding (form) | `64px 48px 52px` |
| Column gap (two-col layout) | `64px` |
| Button stack gap | `30px` |
| Sidebar nav item gap | `32px` |
| Logo size | `86×86px` |
| Header height | `64px` (h-16) |
| Sidebar width | `80px` |
| Max card width (auth) | `988px` |
| Max card width (onboarding) | `800px` |
| Max form inner width | `550–680px` |

---

## 4. Border Radius

| Element | Radius |
|---|---|
| Cards / modals | `42px` |
| Primary / pill buttons | `50px` |
| Continue button | `24px` |
| Inputs (standard) | `rounded-xl` (12px) |
| Inputs (floating label) | `rounded-lg` (8px) |
| Social icon buttons | `6px` |
| Header search input | `rounded-lg` (8px) |
| Sidebar nav (active) | implicit via padding |
| Loader spinner | `rounded-full` |
| Avatar | `rounded-full` |

---

## 5. Shadows & Elevation

No box-shadows are used. Elevation is communicated purely through borders:

- Auth cards: `border: 1px solid rgba(0,0,0,0.12)`
- Onboarding cards: `border: 1px solid #8E8E93` (gray-400 equivalent)
- Dashboard header/sidebar: `border-b border-gray-200` / `border-r border-gray-200`
- Inputs default: `border border-gray-300`
- Inputs focused: `focus:border-blue-500` or `focus:border-blue-600`
- Inputs error: `border-red-400 focus:border-red-500`

---

## 6. Components

### Button (`src/components/ui/Button.svelte`)

Three variants, all share: `inline-flex items-center justify-center font-medium transition-colors`

| Variant | Background | Text | Border |
|---|---|---|---|
| `primary` | `#2563EB` → hover `#1D4ED8` | white | none |
| `outline` | white → hover `blue-50` | `#2563EB` | `2px solid #2563EB` |
| `ghost` | transparent → hover `gray-100` | `gray-700` | none |

Disabled state: `opacity-60 cursor-not-allowed pointer-events-none`

In-page pill buttons (not using the component) follow this pattern:
```
height: 56px  border-radius: 50px  width: ~199px
Selected: bg #EFF6FF, border #BFDBFE, text #2563EB
Default:  bg transparent, border #2563EB, text #2563EB
```

---

### Card (`src/components/ui/Card.svelte`)

```
bg-white  border-radius: 42px  border: 1px solid rgba(0,0,0,0.3)
```
Accepts `className` and `style` props for overrides.

---

### FloatingInput (`src/components/ui/FloatingInput.svelte`)

Animated floating label input. Label transitions from center to top-left on focus or when value is present.

- Height: auto (pt-4 pb-2.5)
- Border: `border-gray-300` → `focus:border-blue-600`
- Error: `border-red-400 focus:border-red-500`, label turns `text-red-500`
- Focused label: `text-blue-600`
- Default label: `text-gray-600`
- Password toggle: eye/eye-off SVG icon, `text-gray-500 hover:text-gray-700`
- Font size: `text-sm` for both input and label

---

### Select (`src/components/ui/Select.svelte`)

```
height: 56px (h-14)  padding: px-4  border-radius: rounded-xl
border-gray-300 → focus:border-blue-500
Empty value: text-gray-400  |  Selected: text-black
```
Custom chevron SVG (`w-4 h-4 text-gray-500`) positioned `absolute right-4`.

---

### Checkbox (`src/components/ui/Checkbox.svelte`)

```
w-5 h-5  rounded  border-2 border-gray-300
checked accent: text-blue-600  focus-ring: blue-500
```
Label: `text-sm text-gray-700`

---

### FieldError (`src/components/ui/FieldError.svelte`)

```
flex items-center gap-1  text-sm text-red-500  mt-0.5
```
Prefixed with a 16×16 info circle SVG icon.

---

### OTPInput (`src/components/OTPInput.svelte`)

- One `<input>` per digit slot, auto-advances on input, auto-retreats on backspace
- Default slot size: `w-16 h-16` (configurable via `sizeClass` prop)
- Error state: `border-red-400`
- Supports paste — fills all slots from clipboard

---

### PhoneInput (`src/components/PhoneInput.svelte`)

- Left: country flag + dial code selector (dropdown with search)
- Right: number input
- Validates digit count per country
- Dropdown: `bg-white border border-gray-200 rounded-xl shadow-lg` with search input inside
- 27 countries pre-loaded, default: India (+91)

---

### Loader (`src/components/Loader.svelte`)

Spinning border circle. Props:

| Prop | Options |
|---|---|
| `size` | `sm` (w-4 h-4) · `md` (w-6 h-6) · `lg` (w-8 h-8) |
| `color` | `blue` (border-blue-600) · `white` · `gray` (border-gray-400) |

All variants: `rounded-full animate-spin border-t-transparent`

---

## 7. Layout Patterns

### Auth Pages (Signin / Signup)

```
min-h-screen bg-white
  └─ flex items-center justify-center p-4
       └─ max-w-[988px] card  (border-radius: 42px)
            ├─ Desktop (lg:): grid grid-cols-2 gap-[64px] padding 72px 42px 48px
            │    ├─ Left: logo + heading + account link
            │    └─ Right: action buttons + divider + social icons
            └─ Mobile: single column, same content stacked
```

Two-column desktop split: left = branding/copy, right = action buttons.

---

### Onboarding Pages

```
min-h-screen bg-white flex items-center justify-center
  └─ max-w-[800px] card  (border-radius: 42px, border: 1px solid #8E8E93)
       └─ flex flex-col items-center  padding: 72px 42px 48px
            ├─ Logo (86×86)
            ├─ Heading (36px, font-medium)
            └─ Form content (max-w-[550–680px])
```

---

### Dashboard Layout

```
min-h-screen bg-gray-50
  ├─ DashboardHeader  (fixed, h-16, bg-white, border-b border-gray-200, z-50)
  │    ├─ Left: hamburger (mobile) + logo + "Visionary" wordmark
  │    ├─ Center: search bar (max-w-2xl, h-10, rounded-lg)
  │    └─ Right: language selector + bell icon + avatar (w-10 h-10 rounded-full bg-blue-500)
  ├─ DashboardSidebar  (fixed, left-0, top-16, w-80px, bg-white, border-r border-gray-200, z-40)
  │    └─ Icon nav: Dashboard · Learn · Ask · Practice
  │         Active: color #2563EB  |  Inactive: color #6b7280
  └─ <slot />  (main content area)
```

Mobile: sidebar slides in via `translateX`, overlay `bg-black bg-opacity-50` covers content.

---

## 8. Interaction States

| State | Visual |
|---|---|
| Button hover (primary) | `#1D4ED8` |
| Button hover (outline) | `bg-blue-50` |
| Button hover (ghost) | `bg-gray-100` |
| Input focus | `border-blue-500` or `border-blue-600` |
| Input error | `border-red-400`, label `text-red-500` |
| Input disabled | `opacity` reduced, `cursor-not-allowed` |
| Link hover | `text-blue-700` |
| Nav item active | `color: #2563EB` |
| Nav item inactive | `color: #6b7280` |
| Social button hover | `bg-gray-50` |
| Category pill selected | `bg-#EFF6FF border-#BFDBFE` |
| Category pill default | `bg-transparent border-#2563EB` |
| Disabled button | `bg-gray-300` / `opacity-60` |

---

## 9. Iconography

All icons are inline SVGs — no icon library dependency for UI icons.  
`@iconify/svelte` is installed but icons in core UI are hand-coded SVGs.

Common icon size: `w-5 h-5` (20px) for header actions, `w-6 h-6` (24px) for sidebar nav.  
Stroke style: `stroke="currentColor" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"`

---

## 10. Responsive Strategy

- Breakpoint: `lg` (1024px) is the only active breakpoint used
- Below `lg`: single-column stacked layout, sidebar hidden behind hamburger
- Above `lg`: two-column auth cards, sidebar always accessible
- Mobile padding: `p-4` on page wrappers
- `min()` CSS function used for fluid widths: `width: min(624px, 100%)`

---

## 11. Design Tokens Quick Reference

```css
/* Tailwind @theme (app.css) */
--font-sans:            'Product Sans', sans-serif;
--color-primary:        #2563EB;
--color-primary-hover:  #1D4ED8;
--color-border:         #D1D5DB;
--color-text-primary:   #111827;
--color-text-secondary: #6B7280;

/* Recurring inline values */
card-radius:     42px
button-radius:   50px  (pill) / 24px (continue)
input-radius:    12px (rounded-xl) / 8px (rounded-lg)
header-height:   64px
sidebar-width:   80px
logo-size:       86px
card-max-w-auth: 988px
card-max-w-form: 800px
```
