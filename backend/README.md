# Backend

This is the backend server for NodeHerder, a home automation and IoT device management system. The backend is built with Go and provides REST API endpoints, WebSocket communication, and MQTT integration for managing IoT devices.

## Technology Stack

- **Go 1.20+** - Programming language
- **MQTT** - Message protocol for IoT device communication
- **WebSocket** - Real-time bidirectional communication
- **BoltDB** - Embedded key-value database for metrics and state
- **Gorilla WebSocket** - WebSocket implementation
- **Logrus** - Structured logging

## Prerequisites

- Go 1.20 or higher
- MQTT broker (e.g., Mosquitto) - for device communication

## Installation

The backend uses Go modules for dependency management. Dependencies will be automatically downloaded when you build or run the application.

To manually download dependencies:

```bash
go mod download
```

## Environment Configuration

The application uses environment variables for configuration. Two environment files are available:

- `.env.development` - Development environment settings
- `.env.production` - Production environment settings

Key environment variables:
- `APP_ENV` - Application environment (`development` or `production`)
- `FRONTEND_BASE_URL` - Frontend URL for CORS (default: `http://localhost:4100`)

Additional configuration may be required for MQTT broker connection, database paths, and other settings.

## Building

Build the backend binary:

```bash
go build -o ./nodeherder main.go
```

Or use the makefile from the project root:

```bash
make build_backend
```

## Running

### Development

Run the backend with default settings:

```bash
./nodeherder
```

Or use the makefile:

```bash
make run_backend
```

### Command-Line Arguments

The backend accepts several command-line arguments:

- `-port` - HTTP server port (default: `4110`)
- `-buildType` - Client build type for serving static files
- `-logLevel` - Logging level: `debug`, `info`, `warn`, `error` (default: `info`)

Example:

```bash
./nodeherder -port 4110 -logLevel debug
```

## API Endpoints

The backend provides REST API endpoints for:

- Device management (list, add, remove, configure)
- Dashboard groups management
- Automation rules
- Metrics collection and retrieval
- Device status and control

### WebSocket

Real-time updates are provided via WebSocket at `/ws` endpoint.

### HTTP Data Collection

Devices can send data via HTTP POST to:
```
POST /api/collect
```

## Project Structure

```
backend/
├── internal/
│   ├── api/           # HTTP API handlers
│   ├── controllers/   # Business logic controllers
│   ├── services/      # Service layer
│   ├── mqtt/          # MQTT client and handlers
│   ├── ws/            # WebSocket handlers
│   ├── automations/   # Automation engine
│   ├── fs/            # File system operations
│   ├── ratelimiter/   # API rate limiting
│   └── hub.go         # Main hub/server
├── models/            # Data models
├── repository/        # Data access layer
├── store/             # State management
├── utils/             # Utility functions
├── testing/           # Test utilities and mocks
├── mocks/             # Mock implementations
├── configs/           # Configuration files
├── go.mod             # Go module definition
├── go.sum             # Go dependencies checksums
└── main.go            # Application entry point
```

## Key Features

- **Device Management**: Register, configure, and manage IoT devices
- **MQTT Integration**: Connect to MQTT broker for device communication
- **Metrics Storage**: Store and query device metrics using BoltDB
- **Automation Engine**: Create rules and automations for devices
- **WebSocket Updates**: Real-time updates to connected clients
- **Rate Limiting**: API rate limiting for security
- **Structured Logging**: Comprehensive logging with rotation

## Testing

Run tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Run tests for a specific package:

```bash
go test ./internal/api
```

## Database

The backend uses BoltDB, an embedded key-value database, for storing:
- Device metrics
- Application state
- Configuration data

Database files are typically stored in the application directory.

## Logging

Logs are written to both console and file:
- Log files are rotated automatically
- Log level can be controlled via `-logLevel` flag
- Structured logging provides detailed context for debugging

## Development

### Code Formatting

Format code using Go's standard formatter:

```bash
go fmt ./...
```

### Linting

Use `golangci-lint` for comprehensive linting:

```bash
golangci-lint run
```

### Hot Reload (Optional)

For development with hot reload, you can use tools like:
- [air](https://github.com/cosmtrek/air)
- [realize](https://github.com/oxequa/realize)

## MQTT Configuration

The backend connects to an MQTT broker for device communication. Ensure your MQTT broker is:
1. Running and accessible
2. Configured with the correct host/port
3. Has appropriate authentication (if required)

Devices communicate through MQTT topics following the standard patterns.

## Troubleshooting

### Port Already in Use

If port 4110 is already in use, specify a different port:
```bash
./nodeherder -port 8080
```

### MQTT Connection Issues

- Verify MQTT broker is running
- Check network connectivity to broker
- Verify credentials and permissions
- Check log output for connection errors

### Database Errors

- Ensure write permissions in the application directory
- Check disk space
- Verify BoltDB files are not corrupted

### Build Errors

Ensure Go modules are up to date:
```bash
go mod tidy
go mod download
```

## Contributing

When contributing to the backend:
1. Format your code: `go fmt ./...`
2. Run tests: `go test ./...`
3. Ensure all tests pass
4. Follow Go best practices and idioms
5. Add tests for new functionality

## Performance

The backend is designed to handle:
- Multiple concurrent device connections
- Real-time data streaming via WebSocket
- High-frequency metric collection
- Automation rule evaluation

For optimal performance:
- Use appropriate log levels in production (`info` or `warn`)
- Monitor BoltDB size and compact if needed
- Configure rate limiting based on your use case
