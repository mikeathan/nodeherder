import { fetchWithAuth } from '@/contracts/api';
import { ESPHOME_DEVICES_PATH } from '@/contracts/esphome';
import { addESPHomeDevice, listESPHomeDevices, removeESPHomeDevice, updateESPHomeDevice } from '@/services/esphome';

jest.mock('@/contracts/api', () => ({ fetchWithAuth: jest.fn() }));
const fetchMock = jest.mocked(fetchWithAuth);
const input = { name: 'Radar', address: '192.168.1.2', port: 6053, encryption_key: 'example-key' };

beforeEach(() => fetchMock.mockReset());

test('submits authenticated enrollment without treating it as connected', async () => {
  fetchMock.mockResolvedValue({ ok: true, json: async () => ({ data: { id: 'device-id' }, error: null }) } as Response);
  await addESPHomeDevice(input);
  expect(fetchMock).toHaveBeenCalledWith(ESPHOME_DEVICES_PATH, expect.objectContaining({
    method: 'POST', headers: { 'X-NodeHerder-Request': '1' }, body: JSON.stringify(input),
  }));
});

test('loads redacted connection status and sends explicit configuration changes', async () => {
  fetchMock.mockResolvedValueOnce({ ok: true, json: async () => ({ data: [{ id: 'id', address: '192.168.1.2', port: 6053, status: 'offline', diagnostic: 'connection_failed' }], error: null }) } as Response);
  expect(await listESPHomeDevices()).toEqual([{ id: 'id', address: '192.168.1.2', port: 6053, name: undefined, status: 'offline', diagnostic: 'connection_failed' }]);
  fetchMock.mockResolvedValueOnce({ ok: true, status: 200 } as Response);
  await updateESPHomeDevice('id', { name: 'Radar', address: '192.168.1.2', port: 6053 });
  expect(fetchMock).toHaveBeenLastCalledWith(`${ESPHOME_DEVICES_PATH}/id`, expect.objectContaining({ method: 'PUT', body: JSON.stringify({ name: 'Radar', address: '192.168.1.2', port: 6053 }) }));
  fetchMock.mockResolvedValueOnce({ ok: true, status: 200 } as Response);
  await removeESPHomeDevice('id');
  expect(fetchMock).toHaveBeenLastCalledWith(`${ESPHOME_DEVICES_PATH}/id`, expect.objectContaining({ method: 'DELETE' }));
});

test('reports duplicate, invalid and ambiguous responses without exposing the key', async () => {
  fetchMock.mockResolvedValueOnce({ ok: false, status: 409 } as Response);
  await expect(addESPHomeDevice(input)).rejects.toThrow('already configured');
  fetchMock.mockResolvedValueOnce({ ok: true, json: async () => ({ data: null, error: null }) } as Response);
  await expect(addESPHomeDevice(input)).rejects.toThrow('Check the device list');
  fetchMock.mockRejectedValueOnce(new TypeError('network failed'));
  await expect(addESPHomeDevice(input)).rejects.toThrow('Check the device list');
});
