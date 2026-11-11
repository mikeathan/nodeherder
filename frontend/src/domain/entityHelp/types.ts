import type { Expose } from '@/types/device';

/** Provenance of a resolved explanation, used to choose its presentation. */
export const EntityHelpKinds = {
  Provider: 'provider',
  Catalog: 'catalog',
} as const;

export type EntityHelpKind = (typeof EntityHelpKinds)[keyof typeof EntityHelpKinds];

/** Optional authoritative reference shown alongside a catalog explanation. */
export type EntityHelpSource = { label: string; url: string };

/** One resolved explanation for an entity; `source` is only meaningful for catalog text. */
export type EntityHelp = {
  text: string;
  kind: EntityHelpKind;
  source?: EntityHelpSource;
};

/**
 * Resolves the best explanation for one expose.
 *
 * Implementations return `null` when they have no trustworthy text, so later sources
 * can answer. They must not invent semantics from ranges or option labels.
 */
export type EntityHelpResolver = (expose: Expose) => EntityHelp | null;
