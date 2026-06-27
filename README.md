The project is now [`LIVE`](https://tomtris--incident-handoff--k66vxjf79gbd.code.run/)


# Handoff `//`

A full-stack **incident handoff** tool: a Go HTTP backend plus a Vue 3 frontend that turn the messy,
under-pressure moment of passing an active incident from one engineer to another into a clean,
structured brief.

---

## What problem does this solve?

In teams practicing **YBIYRI** ("You Build It, You Run It"), the engineers who write the code carry
the pager. When an incident outlasts the individual — fatigue after hours of firefighting,
cross-team escalation, timezone handover — one engineer's accumulated context needs to reach another
engineer **intact, under pressure, during an active incident**. In practice it doesn't. It lands in
Slack as fragments, and the next responder burns precious minutes reconstructing *what already
happened* before they can help.

**Handoff** fixes the handoff itself:

- It captures **timestamped, typed actions** as the engineer works — every note is an `observation`,
  `action`, `discovery`, `open_question`, or `state_change`, so the history stays skimmable instead
  of becoming a wall of chat.
- It generates a **structured brief for the next person**: what was done, what's still open, and
  where to start — surfaced as a "catch-up" panel and handoff stats (entries, actions, open
  questions, and how many times the pager has changed hands).
- It tracks **pager ownership** (`on_call`) per incident, so handing off is an explicit, recorded act.
- It is **role-aware**: only an admin or the current on-call engineer can mutate an incident;
  everyone else gets a read-only view.

In short, the repo turns ad-hoc incident chatter into a shareable, role-aware, structured record
optimized for the moment one person takes over from another.

### Progress
- [x] Phase 1 — Production Go HTTP Service
- [x] Phase 2 — Database Integration
- [x] Phase 3 — WebSocket & Real-Time
- [x] Phase 4 — Observability & Feature Flags
- [x] Phase 5 — Authentication
- [x] Phase 5.5 — Testing Backend
- [x] Phase 6 — TypeScript + Vue.js
- [ ] Phase 7 — Full Handoff Frontend
- [ ] Phase 8 — Testing Frontend
- [ ] Phase 9 — Ship It

### How the pieces fit together

- **Layered HTTP.** Every request passes through `RequestID → Observability → Timeout` middleware.
  Routes are grouped into three mounts in `router.go`:
  - `/api/*` — authenticated app routes (incidents, entries, handoff brief, auth/me, incident WebSocket)
  - `/admin/*` — admin-only routes (feature flags, on-call shifts)
  - `/*` — public routes (`/login`, `/healthz`, `/readyz`, `/metrics`) **and the SPA fallback**
- **Auth.** `POST /login` verifies credentials and sets an `HttpOnly`, `Secure`, `SameSite=Strict`
  **JWT cookie** (`access_token`, HS256). `AuthMiddleware` parses it and injects a `UserContext`
  (id, username, role) into the request; `AuthAdminOnlyMiddleware` gates the admin mux.
- **Storage is pluggable.** `IncidentStore` is an interface with **in-memory** and **MongoDB**
  implementations, chosen at boot from `HANDOFF_CONNECT_STRING`; an instrumented decorator records
  store metrics. Users, on-call roster, and feature flags use seeded in-memory stores.
- **Real-time.** Each incident exposes `GET /api/incidents/{id}/ws`; a `Registry`/`Hub` fans timeline
  updates out to connected clients over WebSocket.
- **Observability.** Prometheus metrics at `/metrics`, structured `slog` request logs with request IDs,
  and `/healthz` / `/readyz` (the latter pings MongoDB) for probes.
- **Frontend.** The SPA talks only to the JSON API; in dev, Vite proxies `/api`, `/admin`, `/login`
  to `:8080`, and in production the Go server serves `frontend-vue/dist` with an SPA fallback so
  client-side routes resolve.

### Tech stack

| Layer        | Tech                                                                 |
| ------------ | -------------------------------------------------------------------- |
| Backend      | Go 1.26, `net/http` (std mux), `golang-jwt`, `gorilla/websocket`     |
| Database     | MongoDB 7 (replica set) — or in-memory fallback                      |
| Observability| Prometheus client, `log/slog`                                        |
| Frontend     | Vue 3 + TypeScript, Vite, vue-router, Pinia                          |
| Tests        | Go `testing` (unit + contract); Vitest + Playwright on the frontend  |

---

## Running locally

**Prerequisites:** Go 1.26+, Node 20.19+/22.12+, and Docker (for MongoDB; optional — see below).

```sh
# 1. Configure environment
cp .env.example .env          # then set JWT_SECRET to any non-empty value

# 2. (optional) start MongoDB — omit to use the in-memory store
docker compose up -d db

# 3. Build the frontend and run the server (serves the SPA at http://localhost:8080)
make run
```

For frontend development with hot reload, run the API (`go run .`) and, in another terminal,
`cd frontend-vue && npm install && npm run dev` (Vite proxies API calls to `:8080`).

### Configuration

| Env var                  | Default            | Purpose                                              |
| ------------------------ | ------------------ | ---------------------------------------------------- |
| `JWT_SECRET`             | — (**required**)   | HMAC secret for signing auth tokens                  |
| `HANDOFF_PORT`           | `8080`             | HTTP listen port                                     |
| `HANDOFF_CONNECT_STRING` | `""`               | MongoDB URI; empty → in-memory store                 |
| `HANDOFF_DB`             | `incident_tracker` | MongoDB database name                                |
| `HANDOFF_LOG_LEVEL`      | `info`             | Log level                                            |
| `HANDOFF_ENV`            | `development`      | Environment label                                    |

### Tests

```sh
make test        # go test with HTML coverage report
make test-race   # race detector + coverage
```

### Trial accounts

| username | password   | role     |
| -------- | ---------- | -------- |
| `anh`    | `anh123`   | engineer |
| `bernd`  | `bernd123` | engineer |
| `admin`  | `admin123` | admin    |
