.PHONY: all build build_backend build_frontend run_backend run_frontend run clean

# Default target
all: build

# Backend
build_backend:
	@echo "Cleaning previous backend binary"
	@rm -f ./nodeherder
	@echo "Building node-herder backend"
	@cd backend && go build -o ../nodeherder main.go
	
run_backend: build_backend
	@echo "Running backend"
	@./nodeherder

# Frontend
build_frontend:
	@echo "Cleaning previous frontend dist"
	@rm -rf ./frontend/dist
	@echo "Installing frontend dependencies and building"
	@cd frontend && npm install && npm run build

run_frontend:
	@echo "Running frontend in production mode"
	@cd frontend && npm start

# Combined
build: build_backend build_frontend

run: run_backend run_frontend

# Optional cleanup
clean:
	@echo "Removing built artifacts"
	@rm -f ./nodeherder
	@rm -rf ./frontend/dist
