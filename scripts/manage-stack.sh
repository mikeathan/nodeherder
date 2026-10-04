#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
ACTION=help
STACK=all
DRY_RUN=false
NO_CACHE=false
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
    '  --follow     Follow logs until interrupted (logs only)' \
    '  --tail N     Log lines per service, or all (logs only)' \
    '  --no-color   Disable color (also honors NO_COLOR)' \
    '  -h, --help   Show help without Docker' \
    '' \
    'backend includes backend, mqtt, zigbee2mqtt; frontend includes frontend.' \
    'all uses both Compose files as one project. Run from any directory.' \
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
STEP="$ACTION services"
case "$ACTION" in
  start) run "${COMPOSE[@]}" up -d "${SERVICES[@]}" ;;
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
