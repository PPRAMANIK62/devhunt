# web/CLAUDE.md

The frontend is a thin UI wrapper over the Go API. All business logic, validation, and auth enforcement live in the backend. The frontend's job is to call the API and render the result.

## Commands

```bash
bun dev              # dev server on :5173 (proxies /api/v1 → :8080)
bun run build        # tsc + vite build → dist/
bun run lint         # oxlint
bun run format       # oxfmt --write
```

## Architecture

```
src/
  lib/api.ts          # single fetch wrapper — all API calls go through here
  lib/auth.ts         # JWT read/write in localStorage
  context/            # AuthContext — token + decoded role/userId
  hooks/              # one file per resource (use-jobs, use-company, use-applications)
  pages/              # one file per route
  components/         # layout/, jobs/, company/, applications/, ui/ (shadcn)
  types/index.ts      # mirrors backend models exactly
```

## Patterns to follow

**API calls:** All requests go through `api.get/post/patch/delete` in `lib/api.ts`. Never call `fetch` directly. The wrapper attaches the JWT, unwraps `{ data }`, and throws `ApiError` on non-2xx.

**Data-fetching hooks:** Each resource has a hook in `src/hooks/`. Query hooks (`useJobs`, `useJob`, etc.) return `{ data, loading, error }`. Mutation hooks (`useCreateJob`, etc.) return `{ execute, loading, error }`. All hooks use `AbortController` — create one in `useEffect` and return `() => controller.abort()` as cleanup. For hooks that expose a `refetch`, use `useRef<AbortController>` and abort the previous request before starting a new one.

**`enabled` param:** Hooks that should not fire for unauthenticated or wrong-role users accept an `enabled = true` parameter. Guard the fetch inside `useEffect` with `if (!enabled) return`. See `useMyApplications`.

**Route guards:** Three guard components in `app.tsx` — `GuestRoute` (redirect authenticated users away), `AuthRoute` (redirect unauthenticated), `CompanyRoute` (redirect non-company roles). Use `CompanyRoute` for `/dashboard`. Never rely on a `useEffect` redirect inside a page component as the sole guard.

**Errors:** `ApiError` carries `status` and `code`. Check `err instanceof ApiError && err.status === 404` for not-found handling (see `useMyCompany`). Ignore `AbortError` in catch blocks — it is not a real error.

**No business logic in the frontend.** If something feels like a rule (ownership, role check, status transition), it belongs in Go. The frontend only needs to handle what to show when the API returns 403/404/409.

**Adding a shadcn component:** `bunx shadcn@latest add <component>` from the `web/` directory. Do not overwrite existing files when prompted.
