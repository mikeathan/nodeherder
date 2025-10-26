/* eslint-disable no-console */
import { hubStatePayload } from './state.mjs';
import { createMockToken } from './utils.mjs';

export function registerRoutes(app) {
  // Helpers for mock PKCE/state
  const base64Url = (buf) => buf.toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '');
  const randomString = (len = 32) =>
    base64Url(Buffer.from(Array.from({ length: len }, () => Math.floor(Math.random() * 256))));
  // Register HTTP GET route for /hubstate
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
    res.redirect('http://localhost:4100/?auth=success');
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
}
