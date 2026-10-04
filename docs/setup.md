# Setup and deployment

[Project overview](../README.md) · [API/MCP examples](integrations.md)

## Requirements

For real devices: an Ubuntu/Linux host with Docker Engine and the Compose v2 plugin,
a compatible Zigbee coordinator,
and device access. The setup script uses Bash, OpenSSL, `realpath`, and GNU `sed`;
run it on the Linux host. For native development, use Go from
[`backend/go.mod`](../backend/go.mod) and Node from [`.nvmrc`](../.nvmrc).
UI-only exploration uses the [mock-server path](../README.md#explore-the-ui-without-hardware).

## Configure the host

Create `.env` at the repository root, replacing the coordinator path:

```dotenv
DATA_ROOT=/opt/nodeherder
Z2M_DEVICE=/dev/serial/by-id/your-zigbee-dongle
Z2M_SERIAL_PORT=/dev/ttyUSB0
```

`DATA_ROOT` holds broker/Zigbee2MQTT data outside the repo; it must be writable.
`Z2M_DEVICE` identifies the host device; `Z2M_SERIAL_PORT` is its container path.
Compose rejects missing/empty values rather than silently mounting data under `/mqtt`.

From the repository root:

```bash
./backend/scripts/setup-backend.sh
```

This creates broker directories/config/passwords, writes `MQTT_USER`, `MQTT_PASS`,
and `MQTT_URL=tcp://mqtt:1883` to `backend/.env`, and creates or patches
Zigbee2MQTT's MQTT/serial configuration. Existing Zigbee2MQTT configuration is backed
up before patching. The script regenerates credentials on every run and prints the
password; avoid sharing its output. Restart affected services after credential changes.

## Backend environment

For Compose, create `backend/.env.production`:

```dotenv
APP_ENV=production
FRONTEND_BASE_URL=http://nodeherder.local
FRONTEND_ALLOWED_ORIGINS=http://nodeherder.local
OAUTH_CALLBACK_URL=http://nodeherder.local/api/auth/callback
JWT_SECRET_KEY=<replace-with-random-secret>
GOOGLE_CLIENT_ID=<your-google-client-id>
GOOGLE_CLIENT_SECRET=<your-google-client-secret>
OFFLINE_STRICT_LOCAL=false
```

Use a strong random JWT secret, for example output from `openssl rand -base64 32`.
For HTTPS deployments, replace both URLs with the actual HTTPS origin/callback.

Optional machine-to-machine credentials:

```dotenv
SERVICE_CLIENTS=llm-proxy
SERVICE_SECRET_llm_proxy=<your-service-secret>
```

Client IDs use hyphens if desired; secret variable names replace hyphens with underscores.
Add comma-separated IDs and one secret per client. `FRONTEND_ALLOWED_ORIGINS` must
list allowed browser origins, comma-separated; `FRONTEND_BASE_URL` alone does not
configure CORS. For HTTPS, update the allowed origin too.

For native development, create `backend/.env.development`:

```dotenv
APP_ENV=development
FRONTEND_BASE_URL=http://localhost:4100
FRONTEND_ALLOWED_ORIGINS=http://localhost:4100
OAUTH_CALLBACK_URL=http://localhost:4110/api/auth/callback
JWT_SECRET_KEY=<replace-with-random-secret>
OFFLINE_STRICT_LOCAL=true
```

`APP_ENV` selects development/staging/production; callback URLs support `{PORT}`.
Offline mode chooses the local login flow for requests the backend recognizes as local.
Online login requires Google credentials.

For Google sign-in, create an OAuth client in
[Google Cloud Console](https://console.cloud.google.com/). Set the actual frontend
origin and authorized callback URI; development uses `http://localhost:4100` and
`http://localhost:4110/api/auth/callback`. Keep client credentials on the backend.

## Stack manager

After configuring the host and environment files:

```bash
./scripts/manage-stack.sh --help
./scripts/manage-stack.sh validate all
./scripts/manage-stack.sh rebuild all --dry-run
./scripts/manage-stack.sh rebuild all
./scripts/manage-stack.sh status
./scripts/manage-stack.sh logs backend --tail 50 --follow
```

The script targets Ubuntu/Linux using Bash and standard `dirname`/`sed`; it has no
macOS/Homebrew dependency. Docker/Compose must be accessible to the calling user.
Host Node/Git are optional for deployment; Node from `.nvmrc` is needed for script tests.

| Action | Effect |
| --- | --- |
| `start` | Create/start services; build missing application images |
| `stop` | Stop services, retain containers |
| `down` | Stop/remove selected containers, retain persistent data |
| `restart` | Restart existing containers; no rebuild/configuration update |
| `pull` | Refresh external images; no restart |
| `build` | Build application images; no restart |
| `rebuild` | Pull/build first, then replace changed containers |
| `status` | Show running and stopped containers |
| `logs` | Last 100 lines per service; optional `--tail N\|all`/`--follow` |
| `validate` | Quiet configuration check; no Docker daemon needed |

Scope defaults to `all`; `backend` includes backend, Mosquitto, and Zigbee2MQTT;
`frontend` selects only nginx/frontend. Absolute script paths work from any directory.
Both Compose files combine into one project for `all`; the default project is
`nodeherder`, overridable with `COMPOSE_PROJECT_NAME`. Keep the same project name
as existing deployments. Backend scope reads `backend/.env` and
`backend/.env.production`; container `APP_ENV` is production. Frontend-only commands
do not require backend env files, but serving real data requires an accessible backend.

`--dry-run` prints commands without invoking Docker or checking configuration/daemon;
it is a preview, not validation. `validate` avoids printing resolved configuration;
service logs may still contain sensitive information. `--no-color`/`NO_COLOR` disable
terminal colors; redirected output is plain. `--no-cache` applies to build/rebuild.

Partial `down` uses stop/remove, retaining the shared network for the other stack.
Full `down` removes the combined project's containers/network. Neither deletes
volumes, bind-mounted data, images, or orphan containers. Rebuild leaves existing
containers untouched if pull/build fails; replacement can briefly interrupt service.
There is no automatic rollback or readiness guarantee; inspect status/logs afterward.

Build/rebuild embed the frontend version: `APP_VERSION` overrides it, otherwise the
Node helper derives it from Git/package metadata; without host Node, package version
is used. `start` does not inject a version when building missing images. Docker builds
use Node 24 and Go 1.25 (compatible with the Go minimum/toolchain in `go.mod`).
Context-specific `.dockerignore` files exclude backend secrets/data and frontend
dependencies/output; production Vite env files remain build inputs. Never put secrets
in browser configuration.

Manual equivalent for starting/building both services (without manager version injection):

```bash
docker compose -f docker-compose.backend.yml -f docker-compose.frontend.yml up -d --build
```

Mocked CLI tests run without Docker/devices: `node --test scripts/manage-stack.test.cjs`.
Real configuration checks use temporary dummy env files and an unreachable daemon:

```bash
NODEHERDER_COMPOSE_TESTS=1 node --test scripts/manage-stack.compose.test.cjs
```

They require a discoverable `docker compose`; alternatively set
`NODEHERDER_COMPOSE_BINARY` to an absolute standalone Compose executable path.
Neither suite starts containers or uses production credentials. CI runs both and
Bash syntax checks on Ubuntu. Real container builds/device startup remain separate
deployment-host checks. Lifecycle choices follow
[Docker Compose](https://docs.docker.com/reference/cli/docker/compose/).

## Persistent data

| Host path | Contents |
| --- | --- |
| `${DATA_ROOT}/mqtt/{config,data,log}/` | Broker configuration, credentials, data, logs |
| `${DATA_ROOT}/zigbee2mqtt-data/` | Zigbee network/configuration |
| `backend/configs/` | Application configuration |
| `backend/data/` | Metrics and assistant history databases |
| `backend/logs/` | Application logs |

Back up configuration and persistent data before migration or destructive changes.

## Run the backend natively

With Mosquitto/Zigbee2MQTT running, stop the container backend to free port 4110:

```bash
docker compose -f docker-compose.backend.yml stop backend
cd backend
go build -o nodeherder .
APP_ENV=development MQTT_URL=tcp://localhost:1883 ./nodeherder -port 4110
```

Build the complete Go package (`.`), not only `main.go`. Environment files load beside
the executable; `go run` uses a temporary executable location. Exported variables
take precedence over files. The MQTT host override is needed because `mqtt` is a
Compose DNS name. Generated user/password still come from `backend/.env`.
Native data/log paths default to `data/` and `logs/`; override with `DATA_DIR`/`LOGS_DIR`.

`docker-compose.backend-dev.yml` defines only the backend, not a complete broker/Zigbee
stack. Do not assume it starts those dependencies.

## Frontend

Create `frontend/.env.development.local`:

```dotenv
VITE_API_BASE_URL=http://localhost:4110/api
VITE_WS_BASE_URL=ws://localhost:4110/ws
```

From `frontend/`, run `npm ci`, then `npm run dev`. Vite serves port 4100;
the configured API base must include `/api`, and WebSocket URL `/ws`.
Vite's dev port is configured in `vite.config.js`, not the old documented `PORT` variable.

For production, configure `frontend/.env.production.local` with actual `https://...` API
and `wss://...` WebSocket URLs, then build:

```bash
docker compose -f docker-compose.frontend.yml up -d --build
```

Alternatively, use `./scripts/manage-stack.sh rebuild frontend` from the repository root.

It serves nginx on host port 80 and proxies `/api/` and `/ws` to
`host.docker.internal:4110`. Vite variables are embedded at build time; changing
container environment alone does not update them. For native static serving,
`npm run build`, then `npm start`, which requires `.env.production` and uses
`PORT` (default 9080); it has no nginx API proxy.

## Native systemd examples

[`docs/service/`](service/info) contains native backend/frontend units, not Compose
management. Adapt their hardcoded `/home/mikeathan/dev/node-herder` paths and user,
and build/install native dependencies first. The older single-service example uses
backend port 9080, unlike the current 4110 default. Do not run native units and
containers against the same ports or databases. The stack manager does not manage
systemd units.

## Deployment and access

| Service | Host port | Interface |
| --- | --- | --- |
| Vite development UI | 4100 | HTTP |
| Backend | 4110 | HTTP/WebSocket/MCP |
| Mosquitto | 1883 | MQTT, published by Compose |
| Zigbee2MQTT UI | 8089 | HTTP |
| Container frontend | 80 | HTTP/nginx |

Use HTTPS/WSS and appropriate reverse-proxy/firewall controls for deployment.
Device context and MCP routes are public in `backend/internal/hub.go`; protected APIs
accept session cookies or bearer tokens. The public offline-start route also exists,
so `OFFLINE_STRICT_LOCAL=false` alone does not establish an internet-safe deployment.
Keep secrets out of git, browser variables, and shared logs. Do not assume every
environment filename is ignored: inspect the diff before committing.

## Troubleshooting

- CORS/login: match frontend origin and callback URLs; include extra origins through
  `FRONTEND_ALLOWED_ORIGINS` (including the main frontend origin). Google callback
  URLs must match the authorized URI.
- MQTT: check `docker compose -f docker-compose.backend.yml logs mqtt backend`;
  native backend uses `localhost`, containers use `mqtt`. Check credentials privately.
- Zigbee2MQTT: check `docker compose -f docker-compose.backend.yml logs zigbee2mqtt`
  and `ls -l /dev/serial/by-id/`; verify the coordinator mapping and device permissions.
  On Linux, grant the deployment user the appropriate serial-device group if needed.
- Empty charts: metrics recording defaults off; enable it for the relevant devices.
- MCP unavailable: enable it in Settings → MCP Server; there is no `--mcp` flag.
- Port conflict: the real backend and frontend mock both use 4110; run one at a time.
