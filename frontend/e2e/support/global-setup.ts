import { request } from '@playwright/test';
// Called after Playwright has started owned servers, and before OAuth or any device mutation.
export default async function verifyFixtures() {
  const api = await request.newContext();
  try {
    for (const [url, fixture] of [
      [`${process.env.E2E_HUB_API ?? 'http://localhost:4110/api'}/__test/health`, 'nodeherder-node-browser'],
      ['http://localhost:4120/api/__test/health', 'nodeherder-go-browser'],
    ]) {
      const response = await api.get(url);
      const body = await response.json();
      if (!response.ok() || body.fixture !== fixture || body.run_id !== process.env.NODEHERDER_E2E_RUN_ID)
        throw new Error(`Refusing browser mutations: ${fixture} identity/ownership was not verified.`);
    }
  } finally {
    await api.dispose();
  }
}
