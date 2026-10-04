# Frontend

This is the frontend application for NodeHerder, a home automation and IoT device management system. The application is built with Vue.js 3, TypeScript, and Vite.

## Technology Stack

- **Vue.js 3** - Progressive JavaScript framework
- **TypeScript** - Type-safe JavaScript
- **Vite** - Fast build tool and development server
- **PrimeVue** - UI component library
- **Vuex** - State management
- **Vue Router** - Client-side routing
- **ApexCharts** - Interactive charts and visualizations
- **Bootstrap** - CSS framework

## Prerequisites

- Node.js from [`.nvmrc`](../.nvmrc)
- npm (comes with Node.js)

## Installation

Install dependencies:

```bash
npm ci
```

## Environment Configuration

The application uses environment variables for configuration. Two environment files are available:

- `.env.development` - Development environment settings
- `.env.production` - Production environment settings

Key environment variables:
- `VITE_API_BASE_URL` - Backend API base including `/api` (for example `http://localhost:4110/api`)
- `VITE_WS_BASE_URL` - WebSocket URL including `/ws` (for example `ws://localhost:4110/ws`)

Set local overrides in `.env.development.local`/`.env.production.local`.
Vite embeds these values at build time. Docker lifecycle/nginx details: [setup guide](../docs/setup.md).

## Development

Start the development server with hot-reload:

```bash
npm run dev
```

The application will be available at `http://localhost:4100`; Vite sets its port in `vite.config.js`.

### Other Development Commands

- **Type checking** (watch mode):
  ```bash
  npm run type-check
  ```

- **Linting**:
  ```bash
  npm run lint
  ```

- **Auto-fix linting issues**:
  ```bash
  npm run lint:fix
  ```

- **Format code**:
  ```bash
  npm run format
  ```

- **Run tests**:
  ```bash
  npm test
  ```

- **Run test server** (mock backend):
  ```bash
  npm run test-server
  ```

## Building for Production

Build the application for production:

```bash
npm run build
```

This will:
1. Run TypeScript type checking
2. Create an optimized production build in the `dist/` directory

### Preview Production Build

Preview the production build locally:

```bash
npm run preview
```

### Run Production Build

To serve the production build:

```bash
npm start
```

## Project Structure

```
frontend/
├── src/
│   ├── components/     # Vue components
│   ├── views/          # Page components
│   ├── store/          # Vuex store modules
│   ├── router/         # Vue Router configuration
│   ├── services/       # API services
│   ├── types/          # TypeScript type definitions
│   ├── utils/          # Utility functions
│   ├── assets/         # Static assets
│   ├── configs/        # Configuration files
│   ├── App.vue         # Root component
│   └── main.ts         # Application entry point
├── public/             # Public static assets
├── scripts/            # Build and development scripts
├── tools/              # Development tools (test server, etc.)
└── package.json        # Project dependencies and scripts
```

## Key Features

- Real-time device monitoring via WebSocket
- Dashboard with device groups
- Device management and configuration
- Automation rules and triggers
- Metrics visualization with charts
- Responsive design

## Browser Support

The application supports modern browsers:
- Chrome/Edge (last 2 versions)
- Firefox (last 2 versions)
- Safari (last 2 versions)

## Troubleshooting

### Port Already in Use

If port 4100 is already in use, pass `npm run dev -- --port 4101` or update `vite.config.js`.
Update backend allowed origins and callback/frontend URLs to match. `npm start` separately
uses `PORT` (default 9080) from `.env.production` for static serving.

### Type Checking Errors

Run type checking separately to see detailed errors:
```bash
npm run type-check
```

### Build Fails

Ensure all dependencies are installed:
```bash
rm -rf node_modules package-lock.json
npm install
```

## Contributing

When contributing to the frontend:
1. Run linting before committing: `npm run lint:fix`
2. Ensure types are correct: `npm run type-check`
3. Format your code: `npm run format`
4. Run tests if applicable: `npm test`
