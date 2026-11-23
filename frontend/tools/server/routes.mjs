/* eslint-disable no-console */
import { hubStatePayload } from './state.mjs';
import { createMockToken } from './utils.mjs';
import { registerHistoryRoute } from './history.mjs';
import { getMetricsForDevice } from './metrics-mock.mjs';
import { registerHomeRoutes } from './home.mjs';
import { resetObservationDecisions, setObservationEnabled } from './ws.mjs';
import { CORS_ORIGIN } from './config.mjs';

export function registerRoutes(app) {
  registerHomeRoutes(app, () => hubStatePayload);
  app.get('/api/__test/health', (_req, res) =>
    res.json({ fixture: 'nodeherder-node-browser', run_id: process.env.NODEHERDER_E2E_RUN_ID ?? 'manual-mock' })
  );
  // The mock server outlives a single spec, so state written by one test would be visible to the
  // next. Tests that mutate state reset here to start from the fixture.
  app.post('/api/__test/reset', (_req, res) => {
    resetObservationDecisions();
    res.json({ reset: true });
  });
  // Let a spec put the fixture into the state the radar has to explain: observation turned off.
  app.post('/api/__test/observation', (req, res) => {
    setObservationEnabled(req.body?.enabled !== false);
    res.json({ observing: req.body?.enabled !== false });
  });
  // The ordinary Node mock exposes the same optional-feature degraded contract. Full Connections
  // browser flows use the production Go API with synthetic provider/temp repositories instead.
  const connectionUnavailable = {
    code: 'service_unavailable',
    message: 'Connections are unavailable in this device mock. Use the isolated Go fixture for connection flows.',
  };
  app.get('/api/integrations/catalog', (_req, res) =>
    res.set('Cache-Control', 'private, no-store').json({
      data: {
        available: false,
        availability_error: connectionUnavailable,
        types: [{ id: 'waste_collection', name: 'Bin collections', description: 'Collection dates for your address.' }],
        providers: [],
      },
      error: null,
    })
  );
  app.use('/api/integrations', (_req, res) =>
    res.status(503).set('Cache-Control', 'private, no-store').json({ data: null, error: connectionUnavailable })
  );
  registerHistoryRoute(app, () => hubStatePayload.devices);
  app.get('/api/auth/mode', (req, res) => {
    res.set('Cache-Control', 'no-store').json({ mode: 'google' });
  });
  // Helpers for mock PKCE/state
  const base64Url = (buf) => buf.toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '');
  const randomString = (len = 32) =>
    base64Url(Buffer.from(Array.from({ length: len }, () => Math.floor(Math.random() * 256))));
  // Register HTTP GET route for /hubstate

  // Recorded spatial history fixture shaped like a real day: one three-minute recording holding
  // most of the movement, a one-minute one, two single-sample ones, and a deliberately short
  // "night" recording so a browser test can watch playback finish it and cross into "morning"
  // without waiting. Stretches are in milliseconds from each recording's start.
  const spatialOffsets = {
    'mock-session-night': [0, 1_000, 2_000],
    // Morning keeps a long pause (AC-11) but few marks, so its short segment still has room for
    // one mark per movement; afternoon is denser than its width allows and is sampled down.
    'mock-session-morning': [0, 20_000, 40_000, 143_000, 144_000],
    'mock-session-afternoon': [0, 3_000, 6_000, 9_000, 12_000, 15_000, 18_000, 21_000],
    'mock-session-evening': [0],
  };
  const spatialBases = (now) => ({
    'mock-session-night': now - 10_800_000,
    'mock-session-morning': now - 7_200_000,
    'mock-session-afternoon': now - 3_600_000,
    'mock-session-evening': now - 900_000,
  });
  // One epoch for the whole process: deriving bases from Date.now() per request made a session's
  // summary and its points disagree by however long the test had been running, so a point could
  // fall outside its recording and the scrubber would clamp to an edge.
  const spatialEpoch = Date.now();
  const spatialSessions = (now) => {
    const bases = spatialBases(now);
    return Object.entries(spatialOffsets).map(([id, offsets]) => ({
      id,
      from: bases[id],
      to: bases[id] + offsets[offsets.length - 1],
      points: offsets.length,
      slots: [1],
      capped: false,
      // Every recorded offset is activity: the recorder only stores on movement, so the stored
      // positions ARE the movement. Marking only episode starts made recordings look empty.
      samples: offsets.map((offset) => bases[id] + offset),
    }));
  };
  const spatialPoints = (now, session) => {
    const base = spatialBases(now)[session];
    const offsets = spatialOffsets[session];
    if (base === undefined || !offsets) return [];
    return offsets.map((offset, index) => ({
      at: base + offset,
      slot: 1,
      x: index * 400,
      y: 2000,
    }));
  };
  app.get('/api/spatial/history', (req, res) => {
    const now = spatialEpoch;
    const mode = String(req.query.mode ?? 'sessions');
    if (mode === 'sessions') {
      return res.json(spatialSessions(now));
    }
    if (mode === 'points') {
      return res.json(spatialPoints(now, String(req.query.session ?? '')));
    }
    return res.json([{ x: 0, y: 2000, count: 2 }]);
  });
  app.get('/api/system/storage', (_req, res) => res.json({
    entries: [
      { label: 'Metrics database', bytes: 1024 * 1024, retentionDays: 10 },
      { label: 'Spatial history', bytes: 64 * 1024, retentionDays: 7 },
      { label: 'Logs', bytes: 2 * 1024 },
    ],
    dataDirBytes: 2 * 1024 * 1024,
    diskFreeBytes: 10 * 1024 ** 3,
    diskTotalBytes: 100 * 1024 ** 3,
    metricsRetentionDays: 10,
    positionsRetentionDays: 7,
  }));
  app.get('/api/hubstate', (req, res) => {
    console.log('hubstate GET request');
    res.json(hubStatePayload);
  });

  // Register HTTP POST route for /auth/login - sets state/PKCE cookies, returns OAuth URL
  app.post('/api/auth/login', (req, res) => {
    console.log('login POST request - issuing mock state + PKCE and returning OAuth URL');

    const state = randomString(16);
    const verifier = randomString(64);

    res.cookie('oauthstate', state, {
      httpOnly: true,
      sameSite: 'lax',
      secure: false,
      maxAge: 5 * 60 * 1000,
      path: '/',
    });

    // Bind PKCE to state
    res.cookie('oauthpkce', `${state}.${verifier}`, {
      httpOnly: true,
      sameSite: 'lax',
      secure: false,
      maxAge: 5 * 60 * 1000,
      path: '/',
    });

    const port = app.get('port') || 4110;
    const mockOAuthUrl = `http://localhost:${port}/api/auth/callback?code=mockcode&state=${encodeURIComponent(state)}`;
    res.json({ url: mockOAuthUrl });
  });

  // Register HTTP GET route for /auth/callback - simulates OAuth provider callback
  app.get('/api/auth/callback', (req, res) => {
    console.log('OAuth callback GET request');
    const code = req.query.code;
    const state = req.query.state;

    const stateCookie = req.cookies?.oauthstate;
    const pkceCookie = req.cookies?.oauthpkce;

    if (!code || !state || !stateCookie || !pkceCookie) {
      console.log('Missing code/state or cookies');
      return res.status(400).json({ error: 'invalid_request' });
    }

    if (stateCookie !== state) {
      console.log('State mismatch');
      return res.status(400).json({ error: 'invalid_state' });
    }

    if (!pkceCookie.startsWith(`${state}.`)) {
      console.log('PKCE cookie not bound to state');
      return res.status(400).json({ error: 'invalid_pkce' });
    }
    const user = {
      id: '123',
      username: 'mockuser',
    };

    const token = createMockToken(user, 1440);
    res.cookie('session', token, {
      httpOnly: true,
      secure: false,
      sameSite: 'lax',
      maxAge: 1440 * 60 * 1000, // 1 day
    });

    // Clear temporary cookies
    res.cookie('oauthstate', '', { httpOnly: true, sameSite: 'lax', secure: false, path: '/', maxAge: 0 });
    res.cookie('oauthpkce', '', { httpOnly: true, sameSite: 'lax', secure: false, path: '/', maxAge: 0 });

    // Redirect back to frontend with success flag
    res.redirect(`${CORS_ORIGIN}/?auth=success`);
  });

  // Register HTTP GET route for /auth/me - returns current user session
  app.get('/api/auth/me', (req, res) => {
    console.log('auth/me GET request');
    const sessionCookie = req.cookies?.session;

    if (sessionCookie) {
      // In a real app, you'd verify the token here
      res.json({
        id: '123',
        username: 'mockuser',
      });
    } else {
      res.status(401).json({ error: 'Not authenticated' });
    }
  });

  // Register HTTP POST route for /auth/logout
  app.post('/api/auth/logout', (req, res) => {
    console.log('logout POST request');
    res.clearCookie('session');
    res.json({ status: 'ok' });
  });

  app.post('/api/automation/trigger', (req, res) => {
    console.log('automation trigger POST request');
    const { triggerName, automationId } = req.body || {};
    console.log('triggerName:', triggerName);
    console.log('automationId:', automationId);
    res.json({ success: true });
  });

  // Register HTTP GET route for /metrics
  app.get('/api/metrics', (req, res) => {
    console.log('metrics GET request');
    const { deviceId, expose, from, to } = req.query;

    if (!deviceId || !expose || !from || !to) {
      return res.status(400).json({
        error: 'Missing required parameters: deviceId, expose, from, to',
      });
    }

    const fromMs = parseInt(from);
    const toMs = parseInt(to);

    const metricsData = getMetricsForDevice(deviceId, expose, fromMs, toMs);

    if (!metricsData) {
      return res.status(404).json({
        error: `No metrics data available for device ${deviceId}, expose ${expose}`,
      });
    }

    // Return in the expected format
    res.json({
      type: 'metrics',
      payload: {
        deviceId: deviceId,
        exposes: [metricsData],
      },
    });
  });
}
