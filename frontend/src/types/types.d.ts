export type KeyyValuePair<T> = { [key: string]: T };

export type ValueOf<T> = T[keyof T];

export type Nullable<T> = T | null | undefined;

export type CapitalizedString = Capitalize<string>; 