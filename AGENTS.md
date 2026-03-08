# AI Agent Instructions

## Project Context
- **Overview:** `k-map` is a greenfield B2B geolocation SaaS that provides OSM-backed forward geocoding, reverse geocoding, and proximity validation as a Google Maps alternative. Source-of-truth product docs live in `docs/BRD.md`, `docs/PRD.md`, and `docs/architecture.md`.
- **Tech Stack:** Backend API in `Go`; geocoder engine is self-hosted `Nominatim`; data stores are `PostgreSQL + PostGIS` and `Redis`; local infrastructure via `Docker Compose`; production target is a container platform or Kubernetes. There is no frontend in MVP unless explicitly requested later.
- **Structure:** Target project layout:
```text
docs/
  BRD.md
  PRD.md
  architecture.md
cmd/
  api/
internal/
  api/
  auth/
  config/
  geocode/
  proximity/
  storage/
  telemetry/
migrations/
deploy/
  compose/
  kubernetes/
scripts/
test/
```
- **Core Flow:** `POST /v1/geocode/forward` resolves address -> coordinates; `POST /v1/geocode/reverse` resolves coordinates -> address; `POST /v1/geocode/proximity` geocodes the target query, computes geodesic distance to the input coordinate, and returns `is_near` based on `threshold_meters`. API authenticates by API key, applies tenant rate limits, optionally checks cache, calls internal Nominatim, normalizes responses, and records usage/metrics.
- **Conventions:** Keep the system minimal and backend-first. Prefer explicit REST contracts, small packages, constructor-based dependency injection, and clear boundary separation between transport, domain, and infrastructure. Default API version prefix is `/v1`. Persist app metadata in Postgres and geodata through Nominatim/PostGIS. Use structured JSON logs, request IDs, and tenant IDs on every request. Do not add a frontend, message queue, or search engine unless the requirement explicitly justifies it.
- **Commands:** The repo is currently documentation-first. As code is scaffolded, standardize on these commands rather than inventing a new workflow:
```bash
docker compose -f deploy/compose/docker-compose.yml up -d
docker compose -f deploy/compose/docker-compose.yml down
go run ./cmd/api
go test ./...
golangci-lint run
```

---

## 0. Product Guardrails
- MVP scope is limited to:
  - forward geocoding
  - reverse geocoding
  - proximity validation
  - API key auth
  - tenant rate limiting and usage tracking
- Do not expand scope into routing, map tiles, autocomplete, rich places, or UI unless the user explicitly asks.
- Proximity semantics in MVP are point-to-point distance against the best geocoded candidate, not rooftop certainty or polygon containment.
- Commercial traffic must use self-hosted geocoding infrastructure. Do not design around public shared Nominatim endpoints.

## 1. Git & Issue Conventions (MANDATORY)
**ALL code changes MUST be driven by a GitHub issue.** Informational requests need no issue.
- **Branching:** `feature/<issue-#>-<desc>`, `fix/<issue-#>-<desc>`, `epic/<name>`.
- **Commits:** `type(#issue): short desc` (Types: feat, fix, refactor, docs, chore).
- **Labels:** Use available labels; `epic:<name>` for epics.
- **NEVER:** Commit `.env`, `node_modules/`, `dist/`, local DBs.
- **NEVER:** Use `git add -A` or `git add .` (always stage specific files explicitly).

## 2. Issue Standards
Create issues via `gh issue create` before coding.
**Required Structure:**
```md
### Description & Requirements
[What, why, technical details, files to modify]
### Dependencies
Requires: #XX | Part of: epic:name
### Acceptance Criteria
- [ ] Specific, verifiable deliverable 1 (No vague "works correctly")
- [ ] `golangci-lint run` passes
```

## 3. Workflow Phases

### Phase 1: Research
- `gh issue view <issue-#>`
- Read dependencies and related epics (`gh issue list --label "epic:<name>"`). Ask user if unclear.
- Read `docs/BRD.md`, `docs/PRD.md`, and `docs/architecture.md` before making architecture-affecting changes.

### Phase 2: Implementation & Branching
**Epic Branching Strategy:**
1. **Epic Branch:** `main` → `epic/<name>` (Main integration point).
2. **Sub-issue Branch:** `epic/<name>` → `feature/<issue-#>-<desc>` (NEVER branch from or merge to `main`).
3. **Standalone Issue:** Branch from `main` → `feature/<issue-#>-<desc>`.

Implement changes, follow conventions, and fix all `golangci-lint run` errors.

### Phase 2.5: Testing (MANDATORY)
Before completing an issue, you MUST test and post a report. Do not complete if tests fail.
- **Backend:** Call API (curl/UI); check status/shape/errors.
- **Frontend:** Interact with UI; check states (loading/error/success).
- **Integration:** Run end-to-end flow.
- **Infra/Config:** Run scripts; verify config application.
- **Geo-specific:** Validate forward geocode, reverse geocode, and near/not-near proximity cases against known fixtures.

**Post this Test Report on the issue (`gh issue comment <issue-#> --body "..."`):**
```md
### Test Report
- **Scope & Cases:** [...]
- **Results:** `golangci-lint run` (✅/❌), [Test Cmd/Flow] (✅/❌)
- **Smoke Test:**[ ] App starts/UI accessible, [ ] Core flow works, [ ] No critical errors
```

### Phase 3: Completion & Merging
**A. Epic Auto-Proceed (Sub-issues):** Do NOT ask for user confirmation between issues. Do NOT close the issue.
1. Post Implementation Summary (files changed, what/why) & Test Report.
2. Update checklist: `gh issue edit <id> --body "<updated body with [x]>"`
3. Stage explicitly, commit (`feat(#id): desc`), and push sub-issue branch.
4. Merge sub-issue into `epic/<name>`.
5. Proceed immediately to next issue in epic.

**B. Standalone Issues & Finalizing Epics:** STOP and WAIT for user confirmation.
1. Complete steps 1-4 above (merging standalone to `main`, or waiting to merge Epic to `main`).
2. Only after explicit user approval: merge epic to `main` (if applicable) and execute `gh issue close <id>`.

## 4. GH CLI Command Reference
Use `gh` CLI exclusively. No web UI or raw API calls.
- `gh issue view <id>` (append `--json body -q '.body'` to read/update checklists)
- `gh issue list --label "..."`
- `gh issue create --title "..." --label "..." --body "..."`
- `gh issue edit <id> --body "..."` (or `--add-label "..."`)
- `gh issue comment <id> --body "..."`
- `gh issue close <id>`
- `gh pr create --title "..." --body "..." --base main`

### 5. Multi-Issue & Epic Workflow

**Epic Branching Strategy (MANDATORY)**
Epics use a dedicated integration branch. **NEVER** merge sub-issues directly to `main`.

- **Hierarchy:** `main` → `epic/<name>` → `feature/<issue-#>-<desc>` (or `fix/...`)

**Step-by-Step Execution:**

**1. Setup & Planning**

- List issues: `gh issue list --label "epic:<name>"`
- Work strictly in dependency order. Reference them (`Depends on #XX`).
- Check if epic branch exists: `git branch -a | grep "epic/"`

**2. Start Epic (If branch doesn't exist)**

```bash
git checkout main && git pull origin main
git checkout -b epic/<name> && git push origin epic/<name>
```

**3. Start Sub-Issue (Always branch from Epic)**

```bash
git checkout epic/<name> && git pull origin epic/<name>
git checkout -b feature/<issue-#>-<desc>
```

**4. Complete Sub-Issue (Auto-proceed)**
_Rule: Do NOT ask for confirmation. Do NOT close the issue._

```bash
# Update issue
gh issue comment <id> --body "## Implementation Summary..."
gh issue edit <id> --body "<updated body with [x] checks>"

# Commit & Push Sub-issue
git add <files>
git commit -m "feat(#<id>): <desc>"
git push origin feature/<issue-#>-<desc>

# Merge to Epic Branch
git checkout epic/<name> && git pull origin epic/<name>
git merge feature/<issue-#>-<desc> && git push origin epic/<name>
```

**5. Complete Epic (User Approval Required)**
_Rule: STOP and WAIT for explicit user approval before executing._

```bash
git checkout main && git pull origin main
git merge epic/<name> && git push origin main
gh issue close <epic-issue-#>
```

## 6. Service Design Rules
- Keep the public API stateless.
- Put business logic in `internal/...` packages, not in HTTP handlers.
- Wrap Nominatim behind an internal adapter so upstream changes do not leak into the public API contract.
- Normalize all external geocoder responses into stable response DTOs.
- Prefer synchronous request/response flows for MVP. Do not add async job infrastructure unless batch workloads are explicitly requested.
- Cache only deterministic, safely repeatable geocoding lookups.
- Treat rate limiting, auth, and telemetry as first-class cross-cutting concerns from the start.

## 7. Data and Operations Rules
- Use `PostgreSQL + PostGIS` for persistent system data and geo-capable storage.
- Use `Redis` for rate limiting and short-lived caching.
- Keep OSM/Nominatim import and refresh automation under `scripts/` or `deploy/`.
- Document any new operational dependency in `docs/architecture.md`.
- If a change affects API shape, update both `docs/PRD.md` and `docs/architecture.md` in the same workstream.
