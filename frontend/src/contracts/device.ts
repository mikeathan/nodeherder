import { Expose, ExposePresets, Device } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";
import { KeyyValuePair, ValueOf } from "@/types/types";

export function getExposeAttribute(expose: Expose, name: string): any {
  return expose.attributes ? expose.attributes[name] : null;
}

export function getExposeProperty(expose: Expose, name: string): any {
  return expose.properties ? expose.properties[name] : null;
}

export function getExposePresets(expose: Expose): any {
  if (expose.presets == undefined) {
    return [];
  }

  return expose.presets;
}

export function getExposeBinaryProperty(expose: Expose): boolean {
  if (expose.properties == null) {
    return false;
  }

  if (expose.data == expose.properties["on"]) {
    return true;
  }
  // if (expose.data == expose.properties["off"]) {
  //   return false;
  // }

  return false;
}

export function getFeatureDevices(devices: Device[]): KeyyValuePair<string> {
  let list: KeyyValuePair<string> = {};
  for (const [key, device] of Object.entries(devices)) {
    for (const [key, expose] of Object.entries(device.exposes)) {
      if (expose.properties != undefined) {
        list[device.friendly_name] = device.id;
        break;
      }
    }
  }

  return list;
}

export function getDeviceFeaturesByType(
  device: Device,
  exposeType: ValueOf<typeof ExposeTypes>
): KeyyValuePair<string> {
  return Object.assign(
    {},
    ...Object.values(device.exposes)
      .filter((f) => f.properties != undefined && f.type == exposeType)
      .map((f) => ({ [f.name]: f.name }))
  );
}

export function hasSupportedExposeBinaryProperties(expose: Expose): boolean {
  if (expose.properties == null) {
    return false;
  }

  if (
    expose.data == expose.properties["on"] ||
    expose.data == expose.properties["off"]
  ) {
    return true;
  }

  return false;
}
