# Repository map

Descriptive snapshot: 2026-10-01. Verify code; this map does not set policy.
Toolchain files supersede older setup versions: [go.mod](../../backend/go.mod),
[`.nvmrc`](../../.nvmrc).

| Area | Paths | Evidence |
| --- | --- | --- |
| Startup/wiring | `backend/main.go`, `backend/internal/hub.go`, `backend/store/` | Store/controller tests |
| Devices/MQTT | `backend/internal/{mqtt,services}/`, `backend/models/devices/` | MQTT/processor/registrar/lifetime tests |
| Automations | `backend/internal/automations/`, `backend/models/automations/` | Engine/trigger/scheduler/operation tests |
| Metrics/storage | `backend/internal/metrics/`, `backend/repository/`, `backend/utils/storage/` | Query/aggregation/filter/storage tests |
| HTTP/auth/events | `backend/internal/{api,auth,ws}/` | Route/middleware/token/event-hub tests |
| MCP/assistant | `backend/internal/mcp/`, `backend/models/assistant/`, `backend/store/assistant.go` | Resolver/intent/transport/assistant tests |
| Browser state/network | `frontend/src/{store,contracts,services}/` | `frontend/src/__tests__/` |
| UI | `frontend/src/{components,composables,router}/` | Source/scoped UI checks |
| Dev/deployment | Compose/Dockerfiles, `scripts/`, `frontend/tools/server/`, `.github/workflows/ci.yml` | Setup/mocks/CI |

Braces group directories, not literal paths. Backend layers: models, repositories/
store, services/controllers, transports. MCP separates transport/server/resources/
intent/resolver/metrics. Frontend: Vue 3, Vuex, Vue Router, PrimeVue, Vite, TypeScript,
ApexCharts.

## Important current behavior

- `DeviceLifetimeService.Update`: events bypass equality deduplication; accepted state
  updates precede debounce. Automation gets changed delta; metrics/WS get filtered
  updates. Seed/availability have separate paths. The
  [flow guide](../../backend/docs/architecture.md) omits details; verify code/tests.
- `backend/internal/hub.go`: device context/MCP routes are public; many API/WS routes
  use protected middleware. NH-03 is an obligation, not proof of existing auth coverage.
  Changes must document actual exposure; hardening requires implementation work.
- Assistant: backend API/store; browser `frontend/src/services/assistant.service.ts`.
  MCP resolves/validates targets/metrics deterministically. Query support does not
  authorize future device actions; specify their authorization/execution.
- CI: `go test ./...`, `npm test`. Select/record extra build/lint/race/integration/
  constitutional checks. Frontend mocks do not prove backend parity.
