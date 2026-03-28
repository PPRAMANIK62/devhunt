# DevHunt — Frontend

React frontend for the DevHunt developer job board. Proxies API requests to the Go backend at `localhost:8080`.

## Tech Stack

- **Vite 7** — Build tool
- **React 19** — UI framework
- **TypeScript 5.9** — Type safety
- **Tailwind CSS v4** — Utility-first CSS (via `@tailwindcss/vite`, no config file)
- **shadcn/ui** — Accessible component library
- **React Router v7** — Client-side routing
- **vaul** — Drawer component (via shadcn Drawer)
- **Sonner** — Toast notifications
- **Playfair Display** — Display/heading font (variable)
- **IBM Plex Sans** — Body font (variable)
- **IBM Plex Mono** — Monospace font (metadata, tags, status badges)
- **oxlint** — Fast linter (replaces ESLint), kebab-case filenames enforced
- **oxfmt** — Fast formatter (replaces Prettier)
- **Husky + lint-staged** — Pre-commit hooks (lint + format staged files)

## Prerequisites

- [Bun](https://bun.sh) — package manager and runtime
- Go backend running on `:8080` (see `../README.md`)

## Getting Started

```bash
cd web
bun install
bun run dev       # http://localhost:5173
```

The Vite dev server proxies `/api/*` → `http://localhost:8080`.

## Available Scripts

| Script         | Command                | Description                         |
| -------------- | ---------------------- | ----------------------------------- |
| `dev`          | `bun run dev`          | Start Vite dev server               |
| `build`        | `bun run build`        | TypeScript check + production build |
| `preview`      | `bun run preview`      | Preview production build            |
| `lint`         | `bun run lint`         | Run oxlint                          |
| `format`       | `bun run format`       | Format all files with oxfmt         |
| `format:check` | `bun run format:check` | Check formatting without writing    |

## Project Structure

```
src/
├── app.tsx                         # Router + route definitions
├── main.tsx                        # Entry point (BrowserRouter, ThemeProvider, AuthProvider)
├── index.css                       # Tailwind imports + custom theme vars (amber accent)
├── types/
│   └── index.ts                    # Shared TypeScript types (Job, Company, Application, User)
├── lib/
│   ├── api.ts                      # Typed fetch client (injects Bearer token, wraps errors)
│   ├── auth.ts                     # Token storage + JWT decode helpers (localStorage)
│   └── utils.ts                    # cn() utility
├── context/
│   └── auth-context.tsx            # AuthProvider + useAuth hook
├── hooks/
│   ├── use-jobs.ts                 # useJobs, useJob, useCreateJob, useUpdateJob, useDeleteJob
│   ├── use-company.ts              # useMyCompany, useCompany, useCreateCompany, useUpdateCompany
│   └── use-applications.ts        # useMyApplications, useApply, useUpdateApplicationStatus
├── components/
│   ├── ui/                         # shadcn components
│   ├── layout/
│   │   ├── header.tsx              # Sticky header with nav + auth menu
│   │   └── layout.tsx              # Shell with header + <Outlet /> + Toaster
│   ├── jobs/
│   │   ├── job-card.tsx            # Job listing card
│   │   └── job-form-drawer.tsx     # Create/edit job (Drawer)
│   ├── applications/
│   │   ├── apply-drawer.tsx        # Apply to job with cover note (Drawer)
│   │   └── application-row.tsx     # Application status row
│   └── company/
│       └── company-form-drawer.tsx # Create/edit company profile (Drawer)
└── pages/
    ├── home-page.tsx               # Paginated public job listings
    ├── job-detail-page.tsx         # Job detail + apply
    ├── login-page.tsx              # Login form
    ├── register-page.tsx           # Register form (seeker / company role)
    ├── dashboard-page.tsx          # Company: manage jobs + applicants
    └── applications-page.tsx       # Seeker: track application statuses
```

## Routes

| Path            | Page                 | Access         |
| --------------- | -------------------- | -------------- |
| `/`             | Job listings         | Public         |
| `/jobs/:id`     | Job detail + apply   | Public         |
| `/login`        | Login                | Guest only     |
| `/register`     | Register             | Guest only     |
| `/dashboard`    | Company dashboard    | Auth required  |
| `/applications` | My applications      | Auth required  |

## Conventions

- File naming: **kebab-case** enforced by oxlint
- Imports: `@/` alias maps to `src/`
- Package manager: **bun** (`packageManager` field set in `package.json`)
- No client-side caching — backend handles caching with Redis
- Auth token stored in `localStorage` under key `devhunt_token`
- Pre-commit: oxlint + oxfmt run automatically on staged files via Husky
- Zed editor: format-on-save configured in `.zed/settings.json`
