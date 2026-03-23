import { describe, expect, test, beforeEach } from '@jest/globals';
import { store } from '@/store/index';
import { default as hubState } from '../../../../../../docs/hub_state.json';
import { Device, Devices, DeviceUpdate } from '@/types/device.d';
import { createAppconfig } from '@/contracts/settings';

const devices = (hubState.payload as any).devices;
const mockConfig = createAppconfig();

describe('test devices module', () => {
  beforeEach(() => {
    store.commit('hub/clear');
  });
  test('test hub/findDevice can load all inserted devices', () => {
    // add them to store
    devices.forEach((device: any) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      store.commit('hub/addDevice', newDevice);
    });

    // assert values
    devices.forEach((device: any) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      const result = store.getters['hub/findDevice'](
        newDevice.id
      ) as Device;
      expect(result).toEqual(newDevice);
    });
  });

  test('test hub/listAllDevices returns a list of all devices', () => {
    // add them to store
    devices.forEach((device: any) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      store.commit('hub/addDevice', newDevice);
    });

    // assert values
    const result = store.getters[
      'hub/listAllDevices'
    ]() as Devices;
    devices.forEach((device: any, idx: number) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      expect(result[idx]).toEqual(newDevice);
    });
    expect(result.length).toBe(devices.length);
  });

  test('test hub/clear removes all devices from store', () => {
    // add them to store
    devices.forEach((device: any) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      store.commit('hub/addDevice', newDevice);
    });

    // assert values
    store.commit('hub/clear');
    const result = store.getters[
      'hub/listAllDevices'
    ]() as Devices;

    expect(result.length).toBe(0);
  });

  test('test hub/init inserts all devices in store', () => {
    // assert values
    store.dispatch('hub/init', { devices, config: mockConfig });
    const result = store.getters[
      'hub/listAllDevices'
    ]() as Devices;

    devices.forEach((device: any, idx: number) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      expect(result[idx]).toEqual(newDevice);
    });
    expect(result.length).toBe(devices.length);
  });

  test('test hub/updateDevice update device properties', () => {
    // add device
    var json = JSON.stringify(devices[0]);
    const newDevice: Device = JSON.parse(json);
    store.commit('hub/addDevice', newDevice);

    // update device
    var timestamp = new Date(
      Date.parse('11/30/2011')
    ).toString();
    const update: any = {
      id: newDevice.id,
      last_seen: timestamp,
      data: {
        temperature: 23.12,
        battery: 88,
        state: 'online',
      },
      properties: {
        availability: 'online',
      },
    } as const;
    store.commit('hub/updateDevice', update);

    // evaluate results
    const result = store.getters['hub/findDevice'](
      newDevice.id
    ) as Device;

    expect(result.exposes['temperature'].data).toBe(23.12);
    expect(result.exposes['battery'].data).toBe(88);
  });

  test('test hub/updateList - update store from a list of existing devices', () => {
    // add devices
    devices.forEach((device: any) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      store.commit('hub/addDevice', newDevice);
    });

    // update 2 devices only
    var json = JSON.stringify(devices[0]);
    const th01: Device = JSON.parse(json);
    th01.properties = th01.properties || {};
    (th01.exposes['temperature'] as any).data = 23.3;
    (th01.exposes['humidity'] as any).data = 65.1;
    (th01.properties as any)['availability'] = 'offline';

    var json = JSON.stringify(devices[4]);
    const atticLight: Device = JSON.parse(json);
    (atticLight.exposes['brightness'] as any).data = 100;
    (atticLight.exposes['color_temp'] as any).data = 467;

    // evaluate
    store.commit('hub/setDevices', [
      th01,
      atticLight,
    ] as Devices);

    var result = store.getters['hub/findDevice'](
      th01.id
    ) as Device;
    expect(result.exposes['temperature'].data).toBe(23.3);
    expect(result.exposes['humidity'].data).toBe(65.1);

    var result = store.getters['hub/findDevice'](
      atticLight.id
    ) as Device;
    expect(result.exposes['brightness'].data).toBe(100);
    expect(result.exposes['color_temp'].data).toBe(467);
  });
});
