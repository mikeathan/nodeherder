/* eslint-disable no-console */
import { hubStatePayload } from './state.mjs';
import { createMockToken } from './utils.mjs';

export function registerRoutes(app) {
  // Register HTTP GET route for /hubstate
  app.get('/api/hubstate', (req, res) => {
    console.log('hubstate GET request');
    res.json(hubStatePayload);
  });

  // Register HTTP POST route for /auth/login - returns OAuth URL
  app.post('/api/auth/login', (req, res) => {
    console.log('login POST request - returning mock OAuth URL');
    // Return a mock OAuth URL that points to our callback endpoint
    const mockOAuthUrl = `http://localhost:${app.get('port') || 4110}/api/auth/callback?mock=true`;
    res.json({
      url: mockOAuthUrl,
    });
  });

  // Register HTTP GET route for /auth/callback - simulates OAuth provider callback
  app.get('/api/auth/callback', (req, res) => {
    console.log('OAuth callback GET request');
    const user = {
      id: '123',
      username: 'mockuser',
    };

    const token = createMockToken(user, 1440);
    res.cookie('sessionId', token, {
      httpOnly: true,
      secure: false,
      sameSite: 'lax',
      maxAge: 1440 * 60 * 1000, // 1 day
    });

    // Redirect back to frontend with success flag
    res.redirect('http://localhost:4100/?auth=success');
  });

  // Register HTTP GET route for /auth/me - returns current user session
  app.get('/api/auth/me', (req, res) => {
    console.log('auth/me GET request');
    const sessionCookie = req.cookies?.sessionId;

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
