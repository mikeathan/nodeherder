import type { IntegrationSourceReference } from '@/integrations/types';

/** Source identity only; integrations own their reference vocabulary. */
export interface HomeDeviceSource {
  kind: 'device';
  device_id: string;
  expose: string;
}
/** Whole-device presence placement; the hub derives the values, so no expose is named. */
export interface HomePresenceSource {
  kind: 'presence';
  device_id: string;
}
export interface HomePresenceOverviewSource {
  kind: 'presence_overview';
}
export type HomeCardSource = HomeDeviceSource | HomePresenceSource | HomePresenceOverviewSource | IntegrationSourceReference;
export type HomeBlockKind = 'group' | 'card';

/** An accepted or proposed placement; source readings remain in their existing stores. */
export interface HomeCardPlacement {
  id: string;
  panel: boolean;
}
export type HomeCard = HomeCardSource & HomeCardPlacement;
export type HomeIntegrationCard = IntegrationSourceReference & HomeCardPlacement;
export type HomePresenceCard = HomePresenceSource & HomeCardPlacement;
export type HomePresenceOverviewCard = HomePresenceOverviewSource & HomeCardPlacement;

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
