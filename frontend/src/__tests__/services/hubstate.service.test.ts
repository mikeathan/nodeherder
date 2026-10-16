import { fetchHubState } from '@/services/hubstate.service';

beforeEach(() => { jest.spyOn(console, 'error').mockImplementation(() => {}); });
afterEach(() => { jest.restoreAllMocks(); });

test('snapshot uses established authenticated HTTP helper and forwards cancellation', async () => {
  const body = { devices: [], config: {} };
  const fetch = jest.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: true, json: async () => body } as Response);
  const controller = new AbortController();
  expect(await fetchHubState(controller.signal)).toEqual(body);
  expect(fetch).toHaveBeenCalledWith('http://localhost:4001/hubstate', expect.objectContaining({
    method: 'GET', credentials: 'include', signal: controller.signal,
  }));
});

test('failed HTTP snapshot is rejected rather than accepted as initialized data', async () => {
  jest.spyOn(globalThis, 'fetch').mockResolvedValue({ ok: false, status: 503 } as Response);
  await expect(fetchHubState()).rejects.toThrow('hubstate fetch failed');
});
