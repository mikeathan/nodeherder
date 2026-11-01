import { fetchWithAuth } from '@/contracts/api';
import {
  ESPHOME_DEVICES_PATH,
  ESPHOME_DISCOVERY_PATH,
  ESPHOME_MESSAGES,
  ESPHOME_MUTATION_HEADERS,
  parseESPHomeCandidates,
  parseESPHomeEnrollments,
  validateESPHomeAdded,
} from '@/contracts/esphome';
import type { ESPHomeCandidate, ESPHomeEnrollment, ESPHomeEnrollmentInput, ESPHomeEnrollmentUpdate } from '@/contracts/esphome';
export type { ESPHomeCandidate, ESPHomeEnrollment, ESPHomeEnrollmentInput, ESPHomeEnrollmentUpdate } from '@/contracts/esphome';

/** Browses for untrusted candidates; selecting one does not enroll it. */
export async function discoverESPHomeDevices(signal?: AbortSignal): Promise<ESPHomeCandidate[]> {
  const response = await fetchWithAuth(ESPHOME_DISCOVERY_PATH, { method: 'POST', headers: ESPHOME_MUTATION_HEADERS, signal });
  if (!response.ok) throw new Error(ESPHOME_MESSAGES.discoveryUnavailable);
  return parseESPHomeCandidates(await response.json());
}

/** Loads redacted ESPHome enrollment and connection status for management. */
export async function listESPHomeDevices(signal?: AbortSignal): Promise<ESPHomeEnrollment[]> {
  const response = await fetchWithAuth(ESPHOME_DEVICES_PATH, { method: 'GET', signal });
  if (!response.ok) throw new Error(ESPHOME_MESSAGES.settingsUnavailable);
  return parseESPHomeEnrollments(await response.json());
}

/** Removes an enrollment once; retained metrics and backups are not deleted. */
export async function removeESPHomeDevice(id: string, signal?: AbortSignal): Promise<void> {
  const response = await fetchWithAuth(`${ESPHOME_DEVICES_PATH}/${encodeURIComponent(id)}`, { method: 'DELETE', headers: ESPHOME_MUTATION_HEADERS, signal });
  if (!response.ok) throw new Error(ESPHOME_MESSAGES.removeFailed);
}

/** Replaces connection settings; omitting the key preserves the stored credential. */
export async function updateESPHomeDevice(id: string, input: ESPHomeEnrollmentUpdate, signal?: AbortSignal): Promise<void> {
  const response = await fetchWithAuth(`${ESPHOME_DEVICES_PATH}/${encodeURIComponent(id)}`, { method: 'PUT', headers: ESPHOME_MUTATION_HEADERS, body: JSON.stringify(input), signal });
  if (response.status === 409) throw new Error('This address and port are already configured.');
  if (response.status === 400) throw new Error(ESPHOME_MESSAGES.invalidInput);
  if (!response.ok) throw new Error(ESPHOME_MESSAGES.saveFailed);
}

/** Persists an ESPHome enrollment; connection and discovery continue in the background. */
export async function addESPHomeDevice(input: ESPHomeEnrollmentInput, signal?: AbortSignal): Promise<void> {
  let response: Response;
  try {
    response = await fetchWithAuth(ESPHOME_DEVICES_PATH, {
      method: 'POST',
      headers: ESPHOME_MUTATION_HEADERS,
      body: JSON.stringify(input),
      signal,
    });
  } catch (error) {
    if (error instanceof Error && error.name === 'AbortError') throw error;
    throw new Error(ESPHOME_MESSAGES.unreachable);
  }
  if (response.ok) { await validateAddedResponse(response); return; }
  throw new Error(addFailureMessage(response.status));
}

async function validateAddedResponse(response: Response): Promise<void> {
  let envelope: unknown;
  try { envelope = await response.json(); }
  catch { throw new Error(ESPHOME_MESSAGES.unreadable); }
  validateESPHomeAdded(envelope);
}

function addFailureMessage(status: number): string {
  if (status === 401) return 'Sign in again to add a device.';
  if (status === 403) return 'Reload NodeHerder and try again.';
  if (status === 409) return ESPHOME_MESSAGES.duplicate;
  if (status === 400) return ESPHOME_MESSAGES.invalidInput;
  return 'The device could not be saved. Check device-provider storage and try again.';
}
