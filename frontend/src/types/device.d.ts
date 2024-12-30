import { KeyValuePair, Nullable } from './types';

export type DeviceMap = KeyValuePair<Device>;
export type Devices = Array<Device>;
export type ExposeAttributes = Nullable<KeyValuePair<any>>;
export type ExposePresets = Nullable<KeyValuePair<any>>;
export type ExposeProperties = Nullable<KeyValuePair<any>>;
export type DeviceProperties = KeyValuePair<any>;

export type DeviceFilter = (
  device: Device,
  expose: Expose
) => boolean;

export type Device = {
  id: string;
  friendly_name: string;
  description: string;
  connection_type: string;
  power_source: string;
  exposes: KeyValuePair<Expose>;
  properties: KeyValuePair<any>;
};

export type ExposeType = keyof ExposeTypes;
export type Expose = {
  name: string;
  description: string;
  unit: string;
  data: any | null;
  type: ExposeType;
  attributes: ExposeAttributes;
  presets: ExposePresets;
  properties: ExposeProperties;
};

export type DeviceUpdate = {
  id: string;
  last_seen: string;
  data: KeyValuePair<any>;
  properties: DeviceProperties;
};
