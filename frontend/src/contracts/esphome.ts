/** ESPHome enrollment path relative to the configured API base URL. */
export const ESPHOME_DEVICES_PATH = '/esphome/devices';

/** ESPHome discovery path relative to the configured API base URL. */
export const ESPHOME_DISCOVERY_PATH = '/esphome/discover';

/** Native API default port shown in onboarding. */
export const ESPHOME_DEFAULT_PORT = 6053;
/** Lowest accepted TCP port. */
export const ESPHOME_PORT_MIN = 1;
/** Highest accepted TCP port. */
export const ESPHOME_PORT_MAX = 65535;
/** Maximum UTF-8 bytes accepted for a display name. */
export const ESPHOME_MAX_NAME_LENGTH = 256;
/** Management dialog refresh interval while open. */
export const ESPHOME_POLL_INTERVAL_MS = 5000;

/** Required non-simple header for cookie-authenticated mutations. */
export const ESPHOME_MUTATION_HEADERS = { 'X-NodeHerder-Request': '1' } as const;

/** Shared user-facing failures; components must not invent duplicate wording. */
export const ESPHOME_MESSAGES = {
  discoveryUnavailable: 'Discovery is unavailable. Enter the address manually.',
  settingsUnavailable: 'ESPHome device settings are unavailable.',
  removeFailed: 'The device could not be removed. Check its connection and try again.',
  invalidInput: 'Check the private IP address, port and encryption key.',
  saveFailed: 'The device settings could not be saved.',
  duplicate: 'This ESPHome address and port are already configured.',
  unreachable: 'NodeHerder could not be reached. Check the device list before trying again.',
  unreadable: 'NodeHerder returned an unreadable response. Check the device list before trying again.',
  loadFailed: 'Unable to load ESPHome settings. Check the connection and try again.',
} as const;

export interface ESPHomeEnrollmentInput {
  name: string;
  address: string;
  port: number;
  encryption_key: string;
}

export interface ESPHomeEnrollment {
  id: string;
  name?: string;
  address: string;
  port: number;
  status?: string;
  diagnostic?: string;
  reconnections?: number;
}

export interface ESPHomeCandidate { name: string; address: string; port: number }

export type ESPHomeEnrollmentUpdate = Omit<ESPHomeEnrollmentInput, 'encryption_key'> & { encryption_key?: string };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function envelopeData(envelope: unknown, message: string): unknown {
  if (!isRecord(envelope) || !('data' in envelope) || envelope.error !== null) throw new Error(message);
  return envelope.data;
}

function arrayData(envelope: unknown, message: string): unknown[] {
  const data = envelopeData(envelope, message);
  if (!Array.isArray(data)) throw new Error(message);
  return data;
}

function parseCandidate(value: unknown): ESPHomeCandidate {
  if (!isRecord(value) || typeof value.name !== 'string' || typeof value.address !== 'string' || typeof value.port !== 'number') {
    throw new Error('Discovery returned an unreadable response.');
  }
  return { name: value.name, address: value.address, port: value.port };
}

function optionalString(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined;
}

function parseEnrollment(value: unknown): ESPHomeEnrollment {
  if (!isRecord(value) || typeof value.id !== 'string' || typeof value.address !== 'string' || typeof value.port !== 'number') {
    throw new Error('NodeHerder returned an unreadable response.');
  }
  return {
    id: value.id, address: value.address, port: value.port,
    name: optionalString(value.name), status: optionalString(value.status), diagnostic: optionalString(value.diagnostic),
    reconnections: typeof value.reconnections === 'number' ? value.reconnections : undefined,
  };
}

/** Validates the discovery envelope and projects untrusted candidates into public fields; throws on malformed data. */
export function parseESPHomeCandidates(envelope: unknown): ESPHomeCandidate[] {
  return arrayData(envelope, 'Discovery returned an unreadable response.').map(parseCandidate);
}

/** Validates enrollment list data; optional diagnostics are narrowed and unknown fields are discarded. */
export function parseESPHomeEnrollments(envelope: unknown): ESPHomeEnrollment[] {
  return arrayData(envelope, 'NodeHerder returned an unreadable response.').map(parseEnrollment);
}

/** Requires a successful enrollment receipt without implying that the device connected. */
export function validateESPHomeAdded(envelope: unknown): void {
  const message = 'NodeHerder returned an unreadable response. Check the device list before trying again.';
  const data = envelopeData(envelope, message);
  if (!isRecord(data) || typeof data.id !== 'string' || !data.id) throw new Error(message);
}
