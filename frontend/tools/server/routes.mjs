/* eslint-disable no-console */
import { hubStatePayload } from './state.mjs';

export function registerRoutes(app) {
  // Register HTTP GET route for /hubstate
  app.get('/api/hubstate', (req, res) => {
    console.log('hubstate GET request');
    res.json(hubStatePayload);
  });

  app.post('/api/auth/login', (req, res) => {
    const { username } = req.body;
    // return fake JWT or session token
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
