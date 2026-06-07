# Go + Vue.js Full-Stack Engineering Curriculum

## What is this folder?
A challenge-based curriculum for building production-grade full-stack services. You learn by shipping — not by reading slides.

Every phase from Phase 5 onward builds **Handoff** — an on-call incident handoff platform. The curriculum and the project are the same thing. See here [`fullstack-engineering-curriculum.md`](./fullstack-engineering-curriculum.md). Note: This curriculum (from phase 5 onwards is not evaluated yet).

## Repository Structure

```
0.warm-up/
├── go-flashcards/              React app with 26 flashcards across 4 difficulty levels
└── beginner_challenges/        4 progressive Go challenges with solutions

1.go-backend-engineering/         Phases 1–4: interfaces, concurrency, memory & performance

2.handoff/                        Phases 5–14: the full-stack project
├── backend/                    Go API — HTTP, PostgreSQL, WebSocket, auth, metrics, feature flags
├── frontend/                   Vue.js + TypeScript — dashboard, real-time timeline, auth UI
├── k8s/                        Kubernetes manifests
├── docker-compose.yml
├── .github/workflows/ci.yml
└── README.md
```

## Curriculum

| Phase | Challenge(s) | Hours | Key Topics |
|---|---|---|---|
| 0–1 | Warm up && Word Frequency Counter | 10–15h | syntax, slices, structs, maps, error handling, file I/O |
| 2 | Multi-Format Logger + Shape Calculator Fix | 5–9h | interfaces, io.Writer, json, type switch |
| 3 | Log Generator + Log Scanner + Worker Pool | 15–24h | goroutines, channels, WaitGroup, Mutex, context, net/http client |
| 4 | Benchmark Battle | 4–8h | benchmarks, sync.Pool, escape analysis, pprof |
| 5 | Handoff Incident API | 25–30h | net/http, middleware, slog, config, graceful shutdown, structured errors |
| 6 | PostgreSQL Integration | 15–20h | pgx, migrations, pooling, transactions, interface swap |
| 7 | WebSocket & Real-Time | 12–15h | gorilla/websocket, hub pattern, broadcast, readPump/writePump |
| 8 | Observability + Feature Flags | 18–22h | Prometheus, healthz/readyz, counters/histograms, A/B testing |
| 9 | Authentication | 12–15h | JWT, bcrypt, auth middleware, RBAC, context propagation |
| 10 | Code Review + Pressure | 14–18h | bug spotting, goroutine leaks, race conditions, timed implementation |
| 11 | TypeScript + Vue Components | 25–30h | Vue reactivity, Composition API, defineProps/Emits, CSS, components |
| 12 | Full Handoff Frontend | 25–30h | Pinia, API client, Vue Router, WebSocket client, forms, auth UI |
| 13 | Testing | 20–25h | Go table tests, httptest, race detection, Vitest, Vue Test Utils |
| 14 | Ship It | 18–22h | Docker multi-stage, Compose, GitHub Actions, K8s manifests, deploy |
| | **Total (Phases 5–14)** | **185–227h** | |
| | **Total (Phases 1–14)** | **~250–300h** | |

---

## Progress
- [x] Phase 0 — Warm up
- [ ] Phase 1 — Foundations
- [x] Phase 2 — Interfaces & Type System
- [x] Phase 3 — Goroutines & Channels
- [x] Phase 4 — Memory & Performance
- [x] Phase 5 — Production Go HTTP Service
- [x] Phase 6 — PostgreSQL Integration
- [x] Phase 7 — WebSocket & Real-Time
- [x] Phase 8 — Observability & Feature Flags
- [ ] Phase 9 — Authentication
- [ ] Phase 10 — Code Review & Pressure Test
- [ ] Phase 11 — TypeScript + Vue.js
- [ ] Phase 12 — Full Handoff Frontend
- [ ] Phase 13 — Testing
- [ ] Phase 14 — Ship It

## What is Handoff?

In teams practicing YBIYRI ("You Build It, You Run It"), the engineers who write the code carry the pager. When an incident outlasts the individual — fatigue after hours of firefighting, cross-team escalation, timezone handover — one engineer's accumulated context needs to reach another engineer intact, under pressure, during an active incident. In practice it doesn't. It lands in Slack as fragments.

Handoff captures timestamped actions as the engineer works and generates a structured brief for the next person: what was done, what's still broken, where to start.

## Learning Source

- Phase 0 fundamentals based on [this video](https://www.youtube.com/watch?v=3lazW_dSXKM)
- Phases 1–14 are original challenge-based curriculum