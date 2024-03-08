import { KeyyValuePair, Nullable } from './types';

export type DeviceMap = KeyyValuePair<Device>;
export type Devices = Array<Device>;
export type ExposeAttributes = Nullable<KeyyValuePair<any>>;
export type ExposePresets = Nullable<KeyyValuePair<any>>;
export type ExposeProperties = Nullable<KeyyValuePair<any>>;
export type DeviceProperties = KeyyValuePair<any>;

export type Device = {
  id: string;
  friendly_name: string;
  description: string;
  connection_type: string;
  power_source: string;
  exposes: KeyyValuePair<Expose>;
  properties: KeyyValuePair<any>;
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
  data: KeyyValuePair<any>;
  properties: DeviceProperties;
};
