#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
ACTION=help
STACK=all
DRY_RUN=false
NO_CACHE=false
CONNECTIONS=false
FOLLOW=false
TAIL=100
TAIL_SET=false
COLOR=true
STEP=initializing
START_TIME=$SECONDS

usage() {
  printf '%s\n' \
    'NodeHerder stack manager' \
    'Usage: scripts/manage-stack.sh <action> [all|backend|frontend] [options]' \
    '' \
    '  start      Create/start services (build missing images)' \
    '  stop       Stop services; keep containers and data' \
    '  down       Remove selected containers; keep persistent data' \
    '  restart    Restart existing containers; does not rebuild' \
    '  pull       Refresh external service images without restarting' \
    '  build      Build application images without restarting' \
    '  rebuild    Pull/build first, then replace changed containers' \
    '  status     Show selected containers, including stopped ones' \
    '  logs       Show selected logs (last 100 lines per service)' \
    '  validate   Validate Compose configuration without a daemon' \
    '' \
    '  --dry-run    Print commands only; no Docker/config/daemon checks' \
    '  --no-cache   Disable build cache (build/rebuild only)' \
    '  --connections Include the optional external keyring mount (backend/all)' \
    '  --follow     Follow logs until interrupted (logs only)' \
    '  --tail N     Log lines per service, or all (logs only)' \
    '  --no-color   Disable color (also honors NO_COLOR)' \
    '  -h, --help   Show help without Docker' \
    '' \
    'backend includes backend, mqtt, zigbee2mqtt; frontend includes frontend.' \
    'all uses both Compose files as one project. Run from any directory.' \
    'start/rebuild also resolve TRUSTED_PROXY_CIDRS for local sign-in.' \
    'Ubuntu/Linux: Bash, dirname, sed, Docker Engine + Compose v2; host Node optional.' \
    'APP_VERSION overrides build version; COMPOSE_PROJECT_NAME overrides project.' \
    'Setup and lifecycle details: docs/setup.md'
}

die() { printf 'Error: %s\n' "$*" >&2; exit 2; }
on_error() {
  local code=$?
  printf '\nFailed while %s (exit %s). Later steps were not run.\n' "$STEP" "$code" >&2
  exit "$code"
}
trap on_error ERR
trap 'printf "\nInterrupted; check status before retrying.\n" >&2; exit 130' INT
trap 'printf "\nTerminated; check status before retrying.\n" >&2; exit 143' TERM

if (($#)); then
  case "$1" in
    help|-h|--help) usage; exit 0 ;;
    --*) die "Expected an action first. Use --help." ;;
    *) ACTION=$1; shift ;;
  esac
fi
if (($#)) && [[ "$1" != -* ]]; then STACK=$1; shift; fi
while (($#)); do
  case "$1" in
    --dry-run) DRY_RUN=true ;;
    --no-cache) NO_CACHE=true ;;
    --connections) CONNECTIONS=true ;;
    --follow) FOLLOW=true ;;
    --no-color) COLOR=false ;;
    --tail)
      (($# >= 2)) || die "--tail needs a nonnegative number or all."
      TAIL=$2; TAIL_SET=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "Unknown argument: $1. Use --help." ;;
  esac
  shift
done

case "$ACTION" in
  help) usage; exit 0 ;;
  start|stop|down|restart|pull|build|rebuild|status|logs|validate) ;;
  *) die "Unknown action: $ACTION. Use --help." ;;
esac
case "$STACK" in all|backend|frontend) ;; *) die "Unknown stack: $STACK." ;; esac
if $CONNECTIONS && [[ "$STACK" == frontend ]]; then die "--connections needs backend or all."; fi
if $NO_CACHE && [[ "$ACTION" != build && "$ACTION" != rebuild ]]; then
  die "--no-cache applies only to build/rebuild."
fi
if { $FOLLOW || $TAIL_SET; } && [[ "$ACTION" != logs ]]; then
  die "--follow and --tail apply only to logs."
fi
[[ "$TAIL" == all || "$TAIL" =~ ^[0-9]+$ ]] || die "Invalid --tail: $TAIL."

CYAN=''
GREEN=''
RESET=''
if $COLOR && [[ -t 1 && -z "${NO_COLOR:-}" && "${TERM:-dumb}" != dumb ]]; then
  CYAN=$'\033[36m'; GREEN=$'\033[32m'; RESET=$'\033[0m'
fi
heading() { printf '\n%s> %s%s\n' "$CYAN" "$*" "$RESET"; }
run() {
  if $DRY_RUN; then
    printf '  '; printf '%q ' "$@"; printf '\n'
  else
    "$@"
  fi
}

# Resolve Compose paths/environment from the repository, never the caller's cwd.
cd -- "$ROOT_DIR"
COMPOSE=(docker compose --project-directory "$ROOT_DIR")
if ! $COLOR || [[ ! -t 1 || -n "${NO_COLOR:-}" ]]; then COMPOSE+=(--ansi never); fi
SERVICES=()
if [[ "$STACK" == backend || "$STACK" == all ]]; then
  COMPOSE+=(-f "$ROOT_DIR/docker-compose.backend.yml")
  SERVICES+=(backend mqtt zigbee2mqtt)
  if $CONNECTIONS; then COMPOSE+=(-f "$ROOT_DIR/docker-compose.connections.yml"); fi
fi
if [[ "$STACK" == frontend || "$STACK" == all ]]; then
  COMPOSE+=(-f "$ROOT_DIR/docker-compose.frontend.yml")
  SERVICES+=(frontend)
fi

# Node is optional on deployment hosts. Keep the existing git-aware helper when available.
VERSION=
if [[ "$ACTION" == build || "$ACTION" == rebuild ]]; then
  VERSION=${APP_VERSION:-}
  if [[ -z "$VERSION" ]]; then
    if command -v node >/dev/null 2>&1; then
      VERSION=$(node "$ROOT_DIR/frontend/scripts/get-version.cjs")
    else
      VERSION=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
        "$ROOT_DIR/frontend/package.json")
    fi
  fi
  [[ "$VERSION" =~ ^[[:alnum:]][[:alnum:].+_-]*$ ]] || die "APP_VERSION must be a single version string."
fi

heading "NodeHerder | $ACTION | $STACK"
if $DRY_RUN; then
  printf 'Preview only; configuration and daemon have not been checked.\n'
else
  command -v docker >/dev/null 2>&1 || die "Docker is missing. Install Docker and Compose v2."
  docker compose version >/dev/null 2>&1 || die "Compose v2 is unavailable; install the docker compose plugin."
  STEP="validating configuration; see docs/setup.md for required environment files"
  "${COMPOSE[@]}" config --quiet
  if [[ "$ACTION" != validate ]]; then
    STEP="checking Docker daemon access"
    docker info --format '{{.ServerVersion}}' >/dev/null
  fi
fi

build_images() {
  local args=(build --pull --build-arg "APP_VERSION=$VERSION")
  if $NO_CACHE; then args+=(--no-cache); fi
  STEP="building application images"
  heading "Build version $VERSION"
  run "${COMPOSE[@]}" "${args[@]}" "${SERVICES[@]}"
}

# Local sign-in needs the backend to trust its proxy peers. Compose Nginx reaches the
# backend through the host, so the observed peer is the network gateway; a
# container-to-container proxy appears as the frontend address. Resolve both per run so
# a clean install needs no host-specific TRUSTED_PROXY_CIDRS. Explicit configuration in
# the shell or backend/.env* always wins.
PROXY_OVERRIDE="$ROOT_DIR/.generated/compose.proxy-allowlist.yml"

proxy_allowlist_configured() {
  [[ -n "${TRUSTED_PROXY_CIDRS:-}" ]] && return 0
  grep -qsE '^[[:space:]]*TRUSTED_PROXY_CIDRS=[^[:space:]]' \
    "$ROOT_DIR/backend/.env" "$ROOT_DIR/backend/.env.production" \
    "$ROOT_DIR/backend/.env.staging" "$ROOT_DIR/backend/.env.development" 2>/dev/null
}

# addressed_cidr prints one address as a single-host prefix, or nothing if it is not one.
addressed_cidr() {
  local address=${1:-}
  [[ "$address" == *[.:]* ]] || return 0
  [[ "$address" =~ ^[0-9A-Fa-f:.]+$ ]] || return 0
  if [[ "$address" == *:* ]]; then
    printf '%s/128' "$address"
  else
    printf '%s/32' "$address"
  fi
}

prepare_proxy_allowlist() {
  [[ " $* " == *" backend "* ]] || return 0
  if $DRY_RUN; then
    printf '  # TRUSTED_PROXY_CIDRS resolved from the Compose network before services start\n'
    return 0
  fi
  heading "Proxy trust"
  if proxy_allowlist_configured; then
    printf 'Using configured TRUSTED_PROXY_CIDRS.\n'
    return 0
  fi
  local project networks network gateway frontend frontend_id cidr peers
  STEP="creating the Compose network to resolve proxy peers"
  "${COMPOSE[@]}" create "${SERVICES[@]}" >/dev/null
  # Ask Compose for its normalized project name: it lowercases and strips characters
  # from directory names, so deriving the network name by hand silently misses.
  project=$("${COMPOSE[@]}" config 2>/dev/null | sed -n 's/^name:[[:space:]]*//p' | sed -n '1p' || true)
  peers=''
  if [[ -n "$project" ]]; then
    STEP="reading the Compose network gateway"
    networks=$(docker network ls --filter "label=com.docker.compose.project=$project" \
      --format '{{.Name}}' 2>/dev/null || true)
    for network in $networks; do
      gateway=$(docker network inspect "$network" \
        --format '{{(index .IPAM.Config 0).Gateway}}' 2>/dev/null | sed -n '1p' || true)
      if [[ "$gateway" == '<no value>' ]]; then gateway=''; fi
      cidr=$(addressed_cidr "$gateway")
      if [[ -n "$cidr" && ",$peers," != *",$cidr,"* ]]; then
        peers="${peers:+$peers,}$cidr"
      fi
    done
  fi
  if [[ " $* " == *" frontend "* ]]; then
    STEP="reading the frontend container address"
    frontend_id=$("${COMPOSE[@]}" ps -aq frontend 2>/dev/null | sed -n '1p' || true)
    if [[ -n "$frontend_id" ]]; then
      frontend=$(docker inspect "$frontend_id" \
        --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{"\n"}}{{end}}' 2>/dev/null | sed -n '1p' || true)
      frontend=${frontend//[[:space:]]/}
      cidr=$(addressed_cidr "$frontend")
      if [[ -n "$cidr" && ",$peers," != *",$cidr,"* ]]; then
        peers="${peers:+$peers,}$cidr"
      fi
    fi
  fi
  if [[ -z "$peers" ]]; then
    printf 'Warning: could not resolve the Compose proxy peer; set TRUSTED_PROXY_CIDRS manually (docs/setup.md).\n' >&2
    return 0
  fi
  mkdir -p "$(dirname "$PROXY_OVERRIDE")"
  printf 'services:\n  backend:\n    environment:\n      TRUSTED_PROXY_CIDRS: "%s"\n' "$peers" \
    >"$PROXY_OVERRIDE"
  COMPOSE+=(-f "$PROXY_OVERRIDE")
  printf 'Trusted proxy peers: %s (written to %s)\n' "$peers" "$PROXY_OVERRIDE"
}
STEP="$ACTION services"
case "$ACTION" in
  start)
    prepare_proxy_allowlist "${SERVICES[@]}"
    STEP="$ACTION services"
    run "${COMPOSE[@]}" up -d "${SERVICES[@]}" ;;
  stop) run "${COMPOSE[@]}" stop "${SERVICES[@]}" ;;
  down)
    if [[ "$STACK" == all ]]; then
      run "${COMPOSE[@]}" down
    else
      # Keep the shared network for the other stack; persistent volumes are never removed.
      run "${COMPOSE[@]}" stop "${SERVICES[@]}"
      run "${COMPOSE[@]}" rm --force "${SERVICES[@]}"
    fi ;;
  restart) run "${COMPOSE[@]}" restart "${SERVICES[@]}" ;;
  pull) run "${COMPOSE[@]}" pull --ignore-buildable "${SERVICES[@]}" ;;
  build) build_images ;;
  rebuild)
    STEP="pulling external service images"
    heading "Pull external images"
    run "${COMPOSE[@]}" pull --ignore-buildable "${SERVICES[@]}"
    build_images
    prepare_proxy_allowlist "${SERVICES[@]}"
    STEP="replacing changed containers"
    heading "Apply built images"
    run "${COMPOSE[@]}" up -d --no-build "${SERVICES[@]}" ;;
  status) run "${COMPOSE[@]}" ps --all "${SERVICES[@]}" ;;
  logs)
    LOG_ARGS=(logs --tail "$TAIL")
    if $FOLLOW; then LOG_ARGS+=(--follow); fi
    run "${COMPOSE[@]}" "${LOG_ARGS[@]}" --timestamps "${SERVICES[@]}" ;;
  validate)
    if $DRY_RUN; then run "${COMPOSE[@]}" config --quiet; fi
    printf 'Configuration %s.\n' "$(if $DRY_RUN; then printf 'check previewed'; else printf valid; fi)" ;;
esac

if $DRY_RUN; then
  printf '\nPreview complete; nothing was executed.\n'
else
  printf '\n%sCompleted%s in %ss.\n' "$GREEN" "$RESET" "$((SECONDS - START_TIME))"
  if [[ "$ACTION" == start || "$ACTION" == rebuild ]]; then
    printf 'Containers started; use status/logs to check application readiness.\n'
  fi
fi
