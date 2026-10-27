import { homeApi, HomeError } from '@/services/home.service';

afterEach(() => jest.restoreAllMocks());

test('save carries auth, CSRF marker, revision and page cancellation; validates returned layout', async () => {
  const layout = { version: 1 as const, revision: 1, etag: '"one"', blocks: [] };
  const fetch = jest
    .spyOn(global, 'fetch')
    .mockResolvedValue(new Response(JSON.stringify({ data: layout, error: null }), { status: 200 }));
  const signal = new AbortController().signal;
  expect(await homeApi.save(layout, signal)).toEqual(layout);
  expect(fetch.mock.calls[0][1]).toMatchObject({
    method: 'PUT',
    signal,
    credentials: 'include',
    headers: expect.objectContaining({ 'If-Match': '"one"', 'X-NodeHerder-Request': '1' }),
  });
});

test('preserves conflict/session errors and treats abort as cancellation', async () => {
  const fetch = jest.spyOn(global, 'fetch');
  fetch.mockResolvedValueOnce(
    new Response(JSON.stringify({ data: null, error: { code: 'revision_conflict', message: 'Reload layout.' } }), {
      status: 409,
    })
  );
  await expect(homeApi.read(new AbortController().signal)).rejects.toMatchObject({
    status: 409,
    failure: { code: 'revision_conflict' },
  });
  fetch.mockResolvedValueOnce(new Response('private body', { status: 401 }));
  await expect(homeApi.read(new AbortController().signal)).rejects.toBeInstanceOf(HomeError);
  const aborted = new DOMException('cancelled', 'AbortError');
  fetch.mockRejectedValueOnce(aborted);
  await expect(homeApi.read(new AbortController().signal)).rejects.toBe(aborted);
});

test.each([null, [], { data: null, error: null }])(
  'rejects malformed success envelopes without publishing state: %p',
  async (envelope) => {
    jest.spyOn(global, 'fetch').mockResolvedValue(new Response(JSON.stringify(envelope), { status: 200 }));
    await expect(homeApi.read(new AbortController().signal)).rejects.toBeInstanceOf(HomeError);
  }
);
