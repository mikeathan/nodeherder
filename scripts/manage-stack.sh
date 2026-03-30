#!/usr/bin/env bash
set -Ee
trap 'echo "❌ Failed at line $LINENO"; exit 1' ERR

ACTION=${1:-help}
STACK=${2:-all}

function print_usage() {
  echo "Usage: ./manage-stack.sh <action> [stack]"
  echo "Actions:"
  echo "  start   - Start containers (docker compose up -d)"
  echo "  stop    - Stop containers (docker compose stop)"
  echo "  down    - Stop and remove containers (docker compose down)"
  echo "  restart - Restart containers (docker compose restart)"
  echo "  pull    - Pull latest images (docker compose pull)"
  echo "  rebuild - Pull latest images, rebuild, and restart containers"
  echo ""
  echo "Stacks:"
  echo "  all     - Default. Manage both backend and frontend"
  echo "  backend - Manage only backend"
  echo "  frontend- Manage only frontend"
}

if [[ "$ACTION" == "help" ]]; then
  print_usage
  exit 0
fi

START_TIME=$(date +%s)

echo "🚀 Executing $ACTION for $STACK stack..."

case $ACTION in
  start)
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml up -d
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml up -d
    ;;
  stop)
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml stop
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml stop
    ;;
  down)
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml down
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml down
    ;;
  restart)
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml restart
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml restart
    ;;
  pull)
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml pull
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml pull
    ;;
  rebuild)
    echo "Stopping containers..."
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml down
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml down

    echo "Getting version..."
    VERSION=$(node frontend/scripts/get-version.cjs)
    echo "✅ Using version: $VERSION"

    echo "Pulling latest images..."
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml pull
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml pull

    echo "Rebuilding images..."
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml build --pull --no-cache --build-arg APP_VERSION=$VERSION
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml build --pull --no-cache --build-arg APP_VERSION=$VERSION

    echo "Starting containers..."
    [[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml up -d
    [[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml up -d
    ;;
  *)
    echo "❌ Unknown action: $ACTION"
    print_usage
    exit 1
    ;;
esac

END_TIME=$(date +%s)
echo "✅ Done."
echo "⏱  Total time: $((END_TIME - START_TIME)) sec"
