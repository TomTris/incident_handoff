# Go + Vue.js Full-Stack Engineering Curriculum

> A challenge-based curriculum for building production-grade full-stack services.
> You learn by shipping — not by reading slides.
> Every phase builds one project from backend to deployment.

---

## Who is this for?

Developers who:
- Know programming fundamentals (any language)
- Want to build and operate production services end-to-end
- Want to work and practice YBIYRI ("You Build It, You Run It")
- Learn by solving real problems, not by following tutorials

## What you'll build

**Handoff** — an on-call incident handoff platform. 

In teams that practice YBIYRI, the engineers who write the code are the same engineers who carry the pager. There is no separate operations team. When the service breaks at 2am, the person who built it is the person who gets woken up. This is the deal: you get full ownership and fast decision-making, but you also get the phone call.

This works well — until the incident outlasts the individual. A SEV1 runs for 9 hours. **Or** after 6, the primary responder is cognitively impaired from fatigue and must hand over to someone fresh. **Or** the platform team spends 90 minutes diagnosing an issue before realizing the root cause is in another team's service — now that team's on-call gets paged and needs everything that was already discovered. **Or** an incident starts at 22:00 in Germany and at 02:00 the engineer can't continue, so a colleague in another timezone takes over. Or, Or, Or

In all three cases, the failure mode is the same: one engineer has built up hours of context — what was observed, what was tried, what failed, what hypothesis remains — and that context needs to reach another engineer intact, under pressure, during an active incident. In practice it doesn't. It lands in Slack as a fragmented summary, or a verbal call where half the details are lost, or a "check Grafana" with no specifics. The incoming engineer spends 30–60 minutes reconstructing what the outgoing engineer already knew.

Handoff solves this. It's an on-call incident handoff platform. As the engineer works an incident, they log timestamped entries — observations, actions taken, discoveries, open questions. When another engineer takes over, they open the incident and see a structured timeline of everything that happened, plus an auto-generated brief that highlights: what was done, what's still broken, and where to start.

This is not a toy project. It exercises every skill a full-stack engineer needs: REST API design, database persistence, real-time communication, authentication, observability, metrics, frontend state management, forms, testing, containerization, CI/CD, and deployment.
## Skill progression

| Milestone | What you can do |
|---|---|
| After Phase 0 | Warm up|
| After Phase 1 | Write basic Go — structs, slices, methods, file I/O |
| After Phase 2 | Design with interfaces, read standard library code |
| After Phase 3 | Build concurrent systems without race conditions |
| After Phase 4 | Profile and optimize Go code for performance |
| After Phase 7 | Build a real-time HTTP service with database persistence |
| After Phase 9 | Secure, instrument, and feature-flag a production Go API |
| After Phase 12 | Build and connect a typed Vue.js frontend with state management, forms, and auth |
| After Phase 14 | Containerize, test, automate, and deploy a full-stack distributed service end-to-end |

## YBIYRI

> *"You Build It, You Run It"* — Werner Vogels, CTO of Amazon Web Services.
> The team that builds a feature owns it in production. No hand-off to ops.
> This curriculum is designed around YBIYRI: you build AND deploy AND test AND monitor everything yourself.

## How to use this curriculum

1. Complete phases in order. Each phase depends on the previous one.
2. Initialize the `handoff` Git repository at the start of Phase 5. Commit after every meaningful milestone.
3. Every challenge has a **Checklist** or **Expected Output** — don't move on until yours matches.
4. When you are stuck for more than 1 hour on the same problem, don't hesitate to read the **Sources** section. If still stuck after another hour, search for the specific error message. Ask for help. Do not skip.

[`fullstack-engineering-curriculum.md`](./fullstack-engineering-curriculum.md)

---

# Full Curriculum Summary

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

*Complete phases in order. Don't skip. Each phase builds on the previous one.*