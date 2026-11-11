import type { Expose } from '@/types/device';
import type { EntityHelpSource } from '@/domain/entityHelp';

/** Scalar values accepted by the existing device command path. */
export type EntityDefaultValue = number | string | boolean;

/** One catalogued factory default with an optional documentation source. */
export type EntityDefault = {
  value: EntityDefaultValue;
  source?: EntityHelpSource;
};

/** One writable expose and the default that should be sent to it. */
export type EntityDefaultTarget = {
  expose: Expose;
  value: EntityDefaultValue;
  source?: EntityHelpSource;
};
