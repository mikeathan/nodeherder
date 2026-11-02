import type { IntegrationSourceReference } from '@/integrations/types';

/** Source identity only; integrations own their reference vocabulary. */
export interface HomeDeviceSource {
  kind: 'device';
  device_id: string;
  expose: string;
}
export type HomeCardSource = HomeDeviceSource | IntegrationSourceReference;
export type HomeBlockKind = 'group' | 'card';

/** An accepted or proposed placement; source readings remain in their existing stores. */
export interface HomeCardPlacement {
  id: string;
  panel: boolean;
}
export type HomeCard = HomeCardSource & HomeCardPlacement;
export type HomeIntegrationCard = IntegrationSourceReference & HomeCardPlacement;

/** Ordered named area, or a standalone block containing exactly one placement. */
export interface HomeBlock {
  id: string;
  kind: HomeBlockKind;
  name: string;
  cards: HomeCard[];
}

/** Revision-fenced private layout DTO. Arrays define order; values/provider payloads are excluded. */
export interface HomeLayout {
  version: 1;
  revision: number;
  etag: string;
  blocks: HomeBlock[];
}

/** Sanitized server failure shown by the page request owner. */
export interface HomeFailure {
  code: string;
  message: string;
}
