import { describe, expect, test } from '@jest/globals';
import { DeviceIconKinds, deviceIconKind } from '@/domain/devices/icon';
import { device, expose } from '../../helpers/devices';

describe('device icon identity', () => {
  test('presence capabilities win over an auxiliary light entity', () => {
    const presence = device('radar', {}, [
      expose('presence', { type: 'binary' as never }),
      expose('led', { type: 'binary' as never, attributes: { entity_kind: 'light' } as never }),
    ]);
    expect(deviceIconKind(presence)).toBe(DeviceIconKinds.Motion);
  });

  test('provider-scoped IDs still use the discovered label for identity', () => {
    const presence = device('radar', {}, [
      expose('00000000-00000001', { type: 'binary' as never, attributes: { label: 'Occupancy' } as never }),
      expose('00000000-00000002', { type: 'binary' as never, attributes: { entity_kind: 'light' } as never }),
    ]);
    expect(deviceIconKind(presence)).toBe(DeviceIconKinds.Motion);
  });

  test('a light-only device still resolves to the light family', () => {
    const light = device('lamp', {}, [expose('state', { type: 'binary' as never, attributes: { entity_kind: 'light' } as never })]);
    expect(deviceIconKind(light)).toBe(DeviceIconKinds.Light);
  });

  test('config and diagnostic support entities do not define identity', () => {
    const radar = device('radar', {}, [
      expose('led', { type: 'binary' as never, category: 'config' as never, attributes: { entity_kind: 'light' } as never }),
      expose('occupancy', { type: 'binary' as never }),
    ]);
    expect(deviceIconKind(radar)).toBe(DeviceIconKinds.Motion);
  });

  test('switch, climate, air and unknown fallbacks remain deterministic', () => {
    expect(deviceIconKind(device('plug', {}, [expose('state', { type: 'binary' as never })]))).toBe(DeviceIconKinds.Power);
    expect(deviceIconKind(device('climate', {}, [expose('temperature')]))).toBe(DeviceIconKinds.Climate);
    expect(deviceIconKind(device('air', {}, [expose('co2')]))).toBe(DeviceIconKinds.Air);
    expect(deviceIconKind(device('mystery', {}, [expose('custom_field')]))).toBe(DeviceIconKinds.Generic);
  });
});
