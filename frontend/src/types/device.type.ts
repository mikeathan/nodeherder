export const ExposeTypes = {
  Empty: '',
  Binary: 'binary',
  Enum: 'enum',
  Numeric: 'numeric',
} as const;

export const ExposeAccessModes = {
  Read: 'read',
  Write: 'write',
  ReadWrite: 'readwrite',
} as const;

export const ExposeCategories = {
  Measurement: 'measurement',
  Diagnostic: 'diagnostic',
  Config: 'config',
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
