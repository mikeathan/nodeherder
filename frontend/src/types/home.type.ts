/** Source identity only. Selection/presentation does not invent a placement ID or panel preference. */
export type HomeCardSource =
  | { kind: 'device'; device_id: string; expose: string }
  | { kind: 'waste_service'; instance_id: string; resource_id: string; service_id: string }
  | { kind: 'waste_summary'; instance_id: string; resource_id: string };

/** An accepted or proposed placement; source readings remain in their existing stores. */
export type HomeCard = HomeCardSource & { id: string; panel: boolean };
export type HomeIntegrationCard = Exclude<HomeCard, { kind: 'device' }>;

/** Ordered named area, or a standalone block containing exactly one placement. */
export interface HomeBlock {
  id: string;
  kind: 'group' | 'card';
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
