import { panelRoom, shortName } from '@/domain/panel';
import { DashboardGroup } from '@/types/settings.type';
import { device, expose } from '../helpers/devices';

const lamp = device('lamp', { friendly_name: 'Kitchen light' }, [
  expose('state', { type: 'binary' as never, access_mode: 'readwrite' as never, values: { on: 'ON', off: 'OFF' } }),
  expose('brightness', { access_mode: 'readwrite' as never }),
]);
const climate = device('th', { friendly_name: 'Kitchen sensor' }, [expose('temperature'), expose('humidity')]);
const door = device('door', { friendly_name: 'Back door' }, [expose('contact', { type: 'binary' as never, description: 'Door' })]);
const lookup = (id: string) => [lamp, climate, door].find((d) => d.id === id);

const group: DashboardGroup = {
  name: 'Kitchen',
  deviceGroup: {
    lamp: { deviceId: 'lamp', exposes: ['brightness', 'state'] },
    th: { deviceId: 'th', exposes: ['temperature', 'humidity'] },
    door: { deviceId: 'door', exposes: ['contact'] },
    gone: { deviceId: 'gone', exposes: ['power'] },
  },
};

describe('panelRoom', () => {
  it('builds a climate dial and one toggle per switchable device', () => {
    const room = panelRoom(group, lookup);
    expect(room.climate).toEqual({ deviceId: 'th', expose: 'temperature' });
    expect(room.humidity).toEqual({ deviceId: 'th', expose: 'humidity' });
    expect(room.buttons).toEqual([
      { deviceId: 'lamp', expose: 'brightness', label: 'Light' },
      { deviceId: 'door', expose: 'contact', label: 'Contact' },
      { deviceId: 'gone', expose: 'power', label: 'Power' },
    ]);
  });

  it('has no dial without a temperature', () => {
    expect(panelRoom({ name: 'Hall', deviceGroup: { door: { deviceId: 'door', exposes: ['contact'] } } }, lookup).climate).toBeNull();
  });
});

describe('shortName', () => {
  it('drops the room prefix only when something remains', () => {
    expect(shortName('Kitchen light', 'Kitchen')).toBe('Light');
    expect(shortName('Kitchen', 'Kitchen')).toBe('Kitchen');
    expect(shortName('Desk lamp', 'Kitchen')).toBe('Desk lamp');
  });
});
