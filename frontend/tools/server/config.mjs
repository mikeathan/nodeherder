/*
 * Mock hub ports are overridable so browser tests can run beside a local dev stack
 * instead of failing on EADDRINUSE. Defaults keep the documented 4110/4100 pair.
 */
export const PORT = Number(process.env.NODEHERDER_MOCK_PORT ?? 4110);
export const CORS_ORIGIN = process.env.NODEHERDER_MOCK_CORS_ORIGIN ?? 'http://localhost:4100';
export const TIMEZONE = 'Europe/London';
export const paths = {
  // Paths are relative to this file's directory (frontend/tools/server)
  hubState: '../../../docs/hub_state.json',
  metrics: {
    light: '../metrics/light.json',
    temperature: '../metrics/temperature.json',
    presence: '../metrics/presence.json',
  },
};
