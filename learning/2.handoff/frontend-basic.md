# Handoff Frontend — Challenge Curriculum (Minimum Slice)

## What is this?

A challenge-based path to a frontend foundation, built against your own Phase 9 Handoff backend. You learn by shipping. Each phase is a challenge with a binary gate — advance only when every box is true.

Purpose: learn the frontend foundations (HTML, CSS, JS, DOM, events, async, fetch, TypeScript). Handoff is the data source those skills act on.

Knowledge lives in the companion file [`frontend-knowledge.md`](./frontend-knowledge.md). This file is the challenges; that file is the textbook. Each phase points to the section to read first.

Vite is deferred to F5 — through F4 you write raw `.html`/`.css`/`.js` and open them in a browser, so you learn the platform without conflating it with tooling.

**Total: 40–60h.**

---

## Step 1 — the minimum slice (this curriculum)

The smallest vertical slice that still exercises every foundation: **log in, then read the incident list.**

| In scope (Step 1) | Concept it forces |
|---|---|
| Login form → POST → receive JWT → hold in memory → send as `Authorization: Bearer` | forms, fetch POST, async, state |
| Authenticated incident list: GET with the token, render it | fetch GET, DOM rendering, loading/error |
| Two views: logged-out (login) and logged-in (list) | conditional rendering from state |

| Deferred to Step 2 | Why |
|---|---|
| Create / resolve incident | adds volume, not new concepts |
| Server filter / sort | additive |
| WebSocket real-time timeline | a whole new transport |
| Routing, brief generation, observability UI | product surface, not foundation |

---

## Backend prerequisites (on you)

1. Backend running on localhost during F4–F6.
2. CORS headers on the backend, or the browser blocks every request. Add a CORS middleware if Phase 9 lacks one. (F4 covers the concept.)
3. Endpoint shapes are yours — substitute your real Phase 9 login path, JWT field, and incident JSON wherever a challenge says "your endpoint." This curriculum invents none.

---

## Curriculum

| Phase | Challenge | Hours | Key Topics | Read first |
|---|---|---|---|---|
| F1 | Static login + incident-list screens | 14–18h | semantic HTML, forms, box model, flexbox/grid, responsive | knowledge §HTML, §CSS |
| F2 | Incident-array logic program (no DOM) | 10–14h | const/let, array methods, closures, modules | knowledge §JavaScript |
| F3 | Interactive list from state (no network) | 8–10h | DOM API, events, render-from-state | knowledge §DOM & events |
| F4 | Wire to backend: login → token → list | 8–10h | fetch, async/await, Bearer auth, CORS, error states | knowledge §Async & fetch |
| F5 | Move to Vite + TypeScript, fully typed | 8–12h | interfaces, unions, narrowing, FetchState, tsconfig, Vite | knowledge §TypeScript |
| F6 | Capstone: the finished minimum slice | 8–10h | integration; **Phase 11 readiness gate** | — |

---

## Challenges

### F1 — Static login + list screens
Build two responsive screens, no JavaScript: (1) a login screen — header, centered form with username/password inputs using native `required` validation, submit button; (2) an incident-list screen — header bar and a list of 4–5 hardcoded incident cards (severity badge, title, timestamp, status).

**Gate**
- [x] Layout uses flexbox/grid, not positioning hacks.
- [x] No overflow or overlap at 375px width.
- [x] Semantic HTML — `<form>`/`<label>`/`<input>` and `<ul>`/`<li>`, not `div` soup.
- [x] You can explain `margin` vs `padding` and `rem` vs `px` unprompted.

### F2 — Incident-array logic
A pure-logic program over a hardcoded array of incident objects (`{ id, severity, status, title, createdAt, resolvedAt }`) shaped like your real incidents: count by severity, count open vs resolved, the list filtered to open, the list sorted by severity then recency. Array methods only. Node or browser console. No DOM.

**Gate**
- [ ] No `var`; no index `for` loop where an array method fits.
- [ ] Empty input and missing fields handled without crashing.
- [ ] You can explain a closure using your own code.

### F3 — Interactive list from state
Make F1's list screen interactive, no framework, using F2's hardcoded array as state (no network yet). Render the list *from* the array. Add a status filter (all/open/resolved) and a sort toggle (severity/recency) that re-render from state.

**Gate**
- [ ] One source of truth: the array. The DOM is rendered from it, never edited as the source.
- [ ] Filter and sort stay consistent with the array.
- [ ] You can state in one sentence what a framework removes here (the manual re-render).

### F4 — Wire to the backend
Replace the hardcoded array with live data. Login form POSTs to your auth endpoint; on success hold the JWT in a JS variable; switch to the list view; GET incidents with `Authorization: Bearer <jwt>`; render them with loading and error states. Use your real endpoint paths and shapes.

**Gate**
- [ ] Login POSTs, receives and stores the token; the list GET sends it as a Bearer header.
- [ ] `fetch` + `await`; `response.ok` checked; failure renders an error state, not a blank page.
- [ ] An unauthenticated list request is rejected and shown; after login it succeeds.

### F5 — Vite + TypeScript
Move the F4 slice into a Vite + TypeScript project. Modules: `render`, `state`, `api`, `auth`. Define an `Incident` interface matching your JSON and an interface for the login response. Type every function including `Promise<Incident[]>`. Model the list's status as `FetchState<Incident[]>`. Enable `strict`.

**Gate**
- [ ] `strict: true`, zero `any`, zero `@ts-ignore`.
- [ ] `Incident` and the login type defined once, imported where needed.
- [ ] The `FetchState` render handles all states exhaustively.
- [ ] Vite dev server runs and hot-reloads.

### F6 — Capstone (Phase 11 readiness gate)
The complete minimum slice in TypeScript on Vite against your backend: login → JWT in memory → authenticated view → GET incidents with Bearer → render with full loading/error. One typed source-of-truth state, four or more modules, responsive. This is Step 1, finished.

**Pass**
- [ ] `tsc --strict --noEmit` (or the Vite build) → zero errors.
- [ ] Deliberate type errors (string for number, missing property, unhandled union variant) are each caught by the compiler.
- [ ] Unauthenticated → rejected and shown; after login → list loads. You can say where the token lives and why.
- [ ] View (login vs list) and rendered list stay consistent with one typed state object.
- [ ] Loading and error states are visible, never silent failure.
- [ ] You can name in 2–3 sentences which parts Vue replaces: manual re-render (reactivity), manual list build (`v-for`), manual event wiring (`@click`), hand-rolled state/token container (Pinia).

Passing F6 unaided → Phase 11 is appropriate. Failing it → the failing box names the blocker; return to that phase.

---

## Step 2 — expansion (after the minimum slice works)

Add one at a time; each is its own read-build-verify loop. Do not start until F6 passes.

| Feature | New concept | Builds on |
|---|---|---|
| Create incident (form → POST) | form state, optimistic vs confirmed updates | F1, F4 |
| Resolve incident (PATCH/PUT) | mutating server state, refetch vs local update | F4 |
| Server filter / sort | query params, input debouncing | F3, F4 |
| Multiple views with URLs | client-side routing | F3 |
| Real-time timeline | WebSocket transport | F4 |
| Brief generation / observability UI | composing the above | all |

Natural decision point: once Step 1's manual DOM work is in your hands, Vue (Phase 11) becomes a concrete convenience for exactly these Step 2 features.

---

## Progress

- [ ] F1 — Static login + list screens
- [ ] F2 — Incident-array logic
- [ ] F3 — Interactive list from state
- [ ] F4 — Wire to the backend
- [ ] F5 — Vite + TypeScript
- [ ] F6 — Capstone → Phase 11 readiness gate