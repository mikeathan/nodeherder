import { KeyValuePair, Nullable } from './types.type';
import { ExposeAccessModes, ExposeCategories, ExposeTypes, DeviceAvailabilityTypes } from './device.type';

export type DeviceMap = KeyValuePair<Device>;
export type Devices = Array<Device>;
export type ExposeAttributes = Nullable<KeyValuePair<any>>;
export type DeviceProperties = KeyValuePair<any>;
export type ExposeValues = Nullable<KeyValuePair<any>>;

export type DeviceFilter = (device: Device, expose: Expose) => boolean;

export type Device = {
  id: string;
  friendly_name: string;
  description: string;
  connection_type: string;
  power_source: string;
  last_seen: string;

  availability: DeviceAvailabilityTypes;
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
  category: ExposeCategory;

  attributes: ExposeAttributes;
  values: ExposeValues;
};

export type DeviceUpdate = {
  id: string;
  last_seen: string;
  availability: ?DeviceAvailabilityTypes;
  data: KeyValuePair<any>;
};
