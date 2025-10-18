/* eslint-disable no-console */
import express from 'express';
import http from 'http';
import cors from 'cors';

import { PORT, CORS_ORIGIN } from './config.mjs';
import { currentTime } from './utils.mjs';
import { initState } from './state.mjs';
import { registerRoutes } from './routes.mjs';
import { registerWebsocket } from './ws.mjs';

// Initialize state from docs and metrics fixtures
initState();

// App and server
const app = express();
const server = http.createServer(app).listen(PORT);
console.log('[' + currentTime() + '] server listening at port ' + PORT);

// Middleware
app.use(
  cors({
    origin: CORS_ORIGIN,
    credentials: true,
  })
);
app.use(express.json());

// Routes and websockets
registerRoutes(app);
registerWebsocket(app, server);
