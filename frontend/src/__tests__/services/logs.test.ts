import { afterEach, describe, expect, test, jest } from '@jest/globals';
import { get, post } from '@/contracts/api';
import { downloadLogFile, listLogFiles, loadLogFile, logFileName } from '@/services/logs';

jest.mock('@/contracts/api', () => ({ get: jest.fn(), post: jest.fn() }));

const mockedGet = get as jest.MockedFunction<typeof get>;
const mockedPost = post as jest.MockedFunction<typeof post>;

afterEach(() => {
  jest.restoreAllMocks();
  mockedGet.mockReset();
  mockedPost.mockReset();
});

describe('log archive service', () => {
  test('lists validated names and rejects a malformed payload', async () => {
    mockedGet.mockResolvedValue({ ok: true, json: async () => ['logs/a.log'] } as Response);
    await expect(listLogFiles()).resolves.toEqual(['logs/a.log']);

    mockedGet.mockResolvedValue({ ok: true, json: async () => [{ name: 'a.log' }] } as unknown as Response);
    await expect(listLogFiles()).rejects.toThrow('unexpected log list');
  });

  test('loads one file, rejects traversal and keeps the display name', async () => {
    mockedPost.mockResolvedValue({ ok: true, headers: new Headers(), text: async () => 'line' } as Response);
    await expect(loadLogFile('logs/nodeherder.log')).resolves.toBe('line');
    expect(mockedPost).toHaveBeenCalledWith('logfile', { file: 'logs/nodeherder.log', action: 'load' });

    await expect(loadLogFile('../secret.log')).rejects.toThrow('not valid');
    expect(mockedPost).toHaveBeenCalledTimes(1);
    expect(logFileName('logs/nodeherder.log')).toBe('nodeherder.log');
  });

  test('download releases the object URL', async () => {
    mockedPost.mockResolvedValue({ ok: true, headers: new Headers(), text: async () => 'log' } as Response);
    jest.spyOn(URL, 'createObjectURL').mockReturnValue('blob:log');
    const revoke = jest.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined);
    jest.spyOn(document, 'createElement').mockReturnValue({ click: jest.fn() } as unknown as HTMLAnchorElement);

    await downloadLogFile('logs/a.log');
    expect(revoke).toHaveBeenCalledWith('blob:log');
  });
});
