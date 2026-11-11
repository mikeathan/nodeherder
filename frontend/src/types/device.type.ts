export const ExposeTypes = {
  Empty: '',
  Binary: 'binary',
  Enum: 'enum',
  Numeric: 'numeric',
} as const;

export const ExposeAccessModes = {
  Read: 'read',
  Readwrite: 'readwrite',
  Write: 'write',
} as const;

export const ExposeCategories = {
  Measurement: 'measurement',
  Diagnostic: 'diagnostic',
  Config: 'config',
} as const;

export const DefaultExposeCategoryList = Object.values(ExposeCategories);

export const DeviceAvailabilityTypes = {
  Unknown: 'unknown',
  Online: 'online',
  Offline: 'offline',
} as const;

/** Connection types emitted by the backend ingress and provider adapters. */
export const ConnectionTypes = {
  Mqtt: 'mqtt',
  Http: 'http',
  ESPHome: 'esphome',
} as const;

/**
 * Shared entity attribute keys written by device adapters.
 * Keep these in sync with the backend provider contract.
 */
export const ExposeAttributeKeys = {
  Label: 'label',
  EntityKind: 'entity_kind',
  ObjectID: 'object_id',
  DeviceClass: 'device_class',
  ControlUnavailable: 'control_unavailable',
  Min: 'min',
  Max: 'max',
  Step: 'step',
  MinLength: 'min_length',
  MaxLength: 'max_length',
} as const;

/** Binary expose value keys used by external device bridges. */
export const ExposeValueKeys = {
  On: 'on',
  Off: 'off',
  Toggle: 'toggle',
} as const;

export const alllowedExposeList: string[] = [
  'temperature',
  'humidity',
  'pressure',
  'presence',
  'illuminance_lux',
  'occupancy',
  'brightness',
  'state',
  'color_temp',
  'air_quality_score',
  'pm1',
  'pm25',
  'pm10',
  'action',
  'action_direction',
  'action_type',
  'action_time',
  'tamper',
  'voc',
  'smoke',
  'smoke_concentration',
  'test',
  'device_fault',
  'power',
  'voltage',
  'current',
  'energy',
];

export type ExposeBinaryColor = {
  on: string;
  off: string;
};
