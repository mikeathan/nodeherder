/**
 * Wireless observation payload types (spec 064).
 *
 * These mirror `backend/models/observations`, which is transport-agnostic: the same shapes
 * carry Bluetooth, WiFi-probe or any future medium. Facets and media are declared by the
 * source, so the viewer never hardcodes knowledge of a provider.
 */

/** Physical medium an observation arrived over. */
export const ObservationMedia = {
  BleAdvertisement: 'ble.adv',
  WiFiProbe: 'wifi.probe',
} as const;

export type ObservationMedium = (typeof ObservationMedia)[keyof typeof ObservationMedia];

/** Operator triage state for one identity. */
export const ObservationDecisions = {
  Unknown: 'unknown',
  Known: 'known',
  Ignored: 'ignored',
} as const;

export type ObservationDecision = (typeof ObservationDecisions)[keyof typeof ObservationDecisions];

/**
 * Delivery counters for one source. They exist to distinguish "connected but silent" from
 * "never configured", so an empty viewer is diagnosable. Process-lifetime only, never history.
 */
export type ObservationSourceDiagnostics = {
  advertisements: number;
  decoded: number;
  dropped: number;
  /** Absent when the source has never delivered an advertisement. */
  lastAdvertisement?: string | null;
};

/** One observing device and the filter dimensions it can populate. */
export type ObservationSourceDescriptor = {
  id: string;
  name: string;
  media: ObservationMedium[];
  facets: string[];
  diagnostics?: ObservationSourceDiagnostics;
};

/** One source's reading of an identity. An absent RSSI means "not observed", not zero. */
export type ObservationSourceReading = {
  id: string;
  rssi?: number | null;
};

/** Aggregated, renderable state of one observed identity. */
export type ObservationIdentityView = {
  key: string;
  kind: string;
  vendor?: number;
  stable: boolean;
  label?: string;
  decision: ObservationDecision;
  lastSeen: string;
  packets: number;
  rssi?: number | null;
  sources: ObservationSourceReading[];
};

/** One complete broadcast snapshot. Never a delta: a missed snapshot reconverges on the next. */
export type ObservationSnapshot = {
  at: string;
  truncated: boolean;
  sources: ObservationSourceDescriptor[];
  identities: ObservationIdentityView[];
  arrivals: number;
};

/** Facet identifiers the viewer understands. Unknown facets are ignored, not rejected. */
export const ObservationFacets = {
  Kind: 'kind',
  Vendor: 'vendor',
  AddressType: 'addressType',
} as const;

export type ObservationFacet = (typeof ObservationFacets)[keyof typeof ObservationFacets];
