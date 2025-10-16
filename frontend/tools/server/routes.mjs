/* eslint-disable no-console */
import { hubStatePayload } from './state.mjs';

export function registerRoutes(app) {
  // Register HTTP GET route for /hubstate
  app.get('/api/hubstate', (req, res) => {
    console.log('hubstate GET request');
    res.json(hubStatePayload);
  });

  app.get('/api/auth/login', (req, res) => {
    console.log('login GET request ', req.query);
    const { username } = req.query || {};
    // return fake JWT or session token

    var token = `mock-jwt-for-${username || 'guest'}`;
    res.json({
      status: 'ok',
      token: `mock-jwt-for-${username || 'guest'}`,
      user: {
        id: '123',
        email: `${username || 'guest'}@example.com`,
      },
    });
  });

  app.post('/api/automation/trigger', (req, res) => {
    console.log('automation trigger POST request');
    const { triggerName, automationId } = req.body || {};
    console.log('triggerName:', triggerName);
    console.log('automationId:', automationId);
    res.json({ success: true });
  });
}
