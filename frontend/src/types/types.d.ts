export type KeyyValuePair<T> = { [key: string]: T };

export type GenericMap<K, V> = {
  [key: K]: V;
};

export type ValueOf<T> = T[keyof T];

export type Nullable<T> = T | null | undefined;

export type CapitalizedString = Capitalize<string>;
