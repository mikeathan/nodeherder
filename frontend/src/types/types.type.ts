export type KeyValuePair<T> = { [key: string]: T };

export type ValueOf<T> = T[keyof T];

export type Nullable<T> = T | null | undefined;

export type CapitalizedString = Capitalize<string>;

export const TimeUnits = {
  UnitMilliseconds: 'milliseconds',
  UnitSeconds: 'seconds',
  UnitMinutes: 'minutes',
  UnitHours: 'hours',
  UnitDays: 'days',
} as const;

export type TimeUnit = ValueOf<typeof TimeUnits>;
export type TimeInterval = {
  value: number;
  unit: TimeUnit;
};
