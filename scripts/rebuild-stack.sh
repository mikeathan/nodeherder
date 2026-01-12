#!/usr/bin/env bash
set -Ee
trap 'echo "❌ Failed at line $LINENO"; exit 1' ERR

START=$(date +%s)

STACK=${1:-all}

echo "Stopping containers..."

[[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml down
[[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml down

echo "Rebuilding images..."

[[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml build --no-cache
[[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml build --no-cache

echo "Starting containers..."

[[ "$STACK" == "backend" || "$STACK" == "all" ]] && docker compose -f docker-compose.backend.yml up -d
[[ "$STACK" == "frontend" || "$STACK" == "all" ]] && docker compose -f docker-compose.frontend.yml up -d

END=$(date +%s)
echo "✅ Done."
echo "⏱  Total time: $((END - START)) sec"