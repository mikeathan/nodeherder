.PHONY: all build build_backend build_frontend run_backend run_frontend run clean

# Default target
all: build

# Backend
build_backend:
	@echo "Cleaning previous backend binary"
	@rm -f ./nodeherder
	@echo "Building node-herder backend"
	@cd backend && go build -o ./nodeherder main.go
	
run_backend: build_backend
	@echo "Running backend"
	@cd backend && ./nodeherder

# Frontend
build_frontend:
	@echo "Cleaning previous frontend dist"
	@rm -rf ./frontend/dist
	@echo "Installing frontend dependencies and building"
	@cd frontend && npm install && npm run build

run_frontend:
	@echo "Running frontend in production mode"
	@cd frontend && npm start

build: build_backend build_frontend

# run:
# 	@echo "Starting backend and frontend in the background"
# 	@parallel ::: "$(MAKE) run_backend" "$(MAKE) run_frontend"

clean:
	@echo "Removing built artifacts"
	@rm -f ./backend/nodeherder
	@rm -rf ./frontend/dist
