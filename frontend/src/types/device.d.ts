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
  provider?: string;
  friendly_name: string;
  description: string;
  connection_type: string;
  power_source: string;
  last_seen: string;

  availability: DeviceAvailabilityTypes;
  exposes: KeyValuePair<Expose>;
  properties: KeyValuePair<any>;
  /**
   * Abstract capabilities this device declares, e.g. "wireless.observe". Consumers gate
   * features on this field rather than on `provider`, so a new source kind needs no UI change.
   * Absent means "capability not declared", never "definitely cannot".
   */
  capabilities?: string[];
};

export type ExposeType = keyof ExposeTypes;
export type ExposeCategory = keyof ExposeCategories;

export type Expose = {
  name: string;
  description: string;
  unit: string;
  data: any | null;
  type: ExposeType;
  category: ExposeCategory;
  access_mode: ExposeAccessModes;
  attributes: ExposeAttributes;
  values: ExposeValues;
};

export type DeviceUpdate = {
  id: string;
  last_seen: string;
  availability: ?DeviceAvailabilityTypes;
  data: KeyValuePair<any>;
};


