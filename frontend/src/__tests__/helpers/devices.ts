import { Device, Expose } from '@/types/device';
import { default as hubState } from '../../../../docs/hub_state.json';

export const hubDevices = (): Device[] => JSON.parse(JSON.stringify((hubState as { payload: { devices: unknown[] } }).payload.devices));
export const hubConfig = () => JSON.parse(JSON.stringify((hubState as { payload: { config: unknown } }).payload.config));

export function expose(name: string, partial: Partial<Expose> = {}): Expose {
  return { name, description: '', unit: '', data: null, type: 'numeric' as never, category: 'measurement' as never, access_mode: 'read' as never, attributes: null, values: null, ...partial } as Expose;
}

export function device(id: string, partial: Partial<Device> = {}, exposes: Expose[] = []): Device {
  return {
    id, friendly_name: id, description: '', connection_type: 'mqtt', power_source: 'mains (single phase)', last_seen: '2026-10-10T10:00:00Z',
    availability: 'online' as never, properties: {}, exposes: Object.fromEntries(exposes.map((e) => [e.name, e])), ...partial,
  } as Device;
}
