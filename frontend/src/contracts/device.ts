import { Expose, DeviceFilter, Device } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { KeyValuePair, ValueOf } from '@/types/types.type';

export function isDeviceOnline(device: Device): boolean {
  return device.availability == 'online';
}

export function getExposeAttribute(expose: Expose, name: string): any {
  return expose.attributes ? expose.attributes[name] : null;
}

export function getExposeProperty(expose: Expose, name: string): any {
  return expose.values ? expose.values[name] : null;
}

export function getExposeBinaryProperty(expose: Expose): boolean {
  if (expose.values == null) {
    return false;
  }

  if (expose.data == expose.values['on']) {
    return true;
  }
  // if (expose.data == expose.properties["off"]) {
  //   return false;
  // }

  return false;
}

export function getDevices(devices: Device[], allowedFilter: DeviceFilter): KeyValuePair<string> {
  let list: KeyValuePair<string> = {};
  for (const [key, device] of Object.entries(devices)) {
    for (const [key, expose] of Object.entries(device.exposes)) {
      if (allowedFilter(device, expose)) {
        list[device.friendly_name] = device.id;
        break;
      }
    }
  }
  return list;
}

export function getPowerSourceValue(device: Device): number {
  if (device.power_source == 'battery') {
    return device.exposes['battery']?.data ?? device.exposes['battpercentage']?.data ?? 0;
  }
  return device.exposes['voltage']?.data ?? 0;
}

export function getExposes(device: Device, filter: DeviceFilter): Array<string> {
  return Object.entries(device.exposes)
    .filter(([id, expose]) => filter(device, expose))
    .map(([i, e]) => e.name);
}

export function getPropertiesByExposeType(device: Device, exposeType: ValueOf<typeof ExposeTypes>): Array<string> {
  return Object.entries(device.exposes)
    .filter(([id, entity]) => entity.type == exposeType)
    .map(([i, e]) => e.name);
}
