/* eslint-disable no-console */
import { hubStatePayload } from './state.mjs';
import { createMockToken } from './utils.mjs';

export function registerRoutes(app) {
  // Register HTTP GET route for /hubstate
  app.get('/api/hubstate', (req, res) => {
    console.log('hubstate GET request');
    res.json(hubStatePayload);
  });

  // Register HTTP POST route for /auth/login
  app.post('/api/auth/login', (req, res) => {
    console.log('login POST request ', req.body);
    const user = {
      id: '123',
      username: 'mockuser',
    };

    const token = createMockToken(user, 1440);
    res.cookie('sessionId', token, {
      httpOnly: true,
      secure: false,
      maxAge: 1440 * 60 * 1000, // 1 day
    });

    res.json({
      status: 'ok',
      token: token,
      user: user,
    });
  });

  // Register HTTP POST route for /auth/logout
  app.post('/api/auth/logout', (req, res) => {
    console.log('logout POST request');

    res.clearCookie('sessionId');
    res.json({ status: 'ok' });
  });

  app.post('/api/automation/trigger', (req, res) => {
    console.log('automation trigger POST request');
    const { triggerName, automationId } = req.body || {};
    console.log('triggerName:', triggerName);
    console.log('automationId:', automationId);
    res.json({ success: true });
  });
}
