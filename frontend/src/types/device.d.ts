import { KeyyValuePair, Nullable } from "./types";

export type DeviceMap = KeyyValuePair<Device>;
export type Devices = Array<Device>;

export type Device = {
  id: string;
  friendly_name: string;
  description: string;
  connection_type: string;
  power_source: string;
  exposes: KeyyValuePair<Expose>;
  properties: KeyyValuePair<any>;
};

export type Expose = {
  name: string;
  description: string;
  unit: string;
  data: any | null;
  type: string | null;
  attributes: Nullable<KeyyValuePair<any>>;
  presets: Nullable<KeyyValuePair<any>>;
  properties: Nullable<KeyyValuePair<any>>;
};

export type DeviceUpdate = {
  id: string;
  last_seen: string;
  data: KeyyValuePair<any>;
  properties: KeyyValuePair<any>;
};
