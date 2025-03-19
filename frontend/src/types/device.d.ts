import { KeyValuePair, Nullable } from './types.type';
import { ExposeAccessModes, ExposeCategories, ExposeTypes } from './expose.type';

export type DeviceMap = KeyValuePair<Device>;
export type Devices = Array<Device>;
export type ExposeAttributes = Nullable<KeyValuePair<any>>;
export type DeviceProperties = KeyValuePair<any>;
export type ExposeValues = Nullable<KeyValuePair<any>>;

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
  access_mode: ExposeAccessMode;
  category: ExposeCategory;

  attributes: ExposeAttributes;
  values: ExposeValues;

};

export type DeviceUpdate = {
  id: string;
  last_seen: string;
  data: KeyValuePair<any>;
  properties: DeviceProperties;
};

{"type":"deviceUpdated","payload":{"id":"0xa4c138e1b5658e68","last_seen":"2025-03-19T15:54:25.090Z","data":{"co2":476,"formaldehyd":26,"pm25":5,"temperature":16.7,"voc":128}}}
