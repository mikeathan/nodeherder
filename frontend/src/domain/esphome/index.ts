/*
 * Pure ESPHome onboarding helpers: input validation, payload construction and
 * user-facing labels. Dialogs keep only presentation and lifecycle concerns.
 */
import type { Device } from '@/types/device';
import {
  ESPHOME_DEFAULT_PORT,
  ESPHOME_PORT_MAX,
  ESPHOME_PORT_MIN,
  type ESPHomeEnrollment,
  type ESPHomeEnrollmentInput,
  type ESPHomeEnrollmentUpdate,
} from '@/contracts/esphome';

/** How the edit form treats the stored encryption key. */
export type ESPHomeKeyMode = 'keep' | 'replace' | 'clear';

/** Add-form validation failure shown before any request is sent. */
export const ESPHOME_ADD_INPUT_ERROR = `Enter a private IP address and a port from ${ESPHOME_PORT_MIN} to ${ESPHOME_PORT_MAX}.`;
/** Edit-form validation failure shown before any request is sent. */
export const ESPHOME_EDIT_INPUT_ERROR = 'Enter a private IP address, valid port and replacement key.';
/** Discovery failure fallback when the request itself cannot be interpreted. */
export const ESPHOME_DISCOVERY_FALLBACK = 'Discovery is unavailable. Enter the address manually.';
/** Success notice for a saved enrollment; it never claims the device is online. */
export const ESPHOME_ADDED_NOTICE = 'ESPHome device added. Connecting and discovering its entities.';
/** Confirmation copy for removing an enrollment. */
export const ESPHOME_REMOVE_CONFIRMATION = {
  message: 'NodeHerder will stop connecting and remove the device. Existing automations, historical metrics and backups are retained.',
  confirmLabel: 'Remove device',
} as const;
/** Fallback label when a saved or discovered device has no name. */
export const ESPHOME_DEFAULT_DEVICE_LABEL = 'ESPHome device';
/** Status shown when a saved enrollment has not reported health yet. */
export const ESPHOME_UNKNOWN_STATUS = 'unknown';
/** Failure after a removal attempt. */
export const ESPHOME_REMOVE_RETRY_FAILED = 'The device could not be removed. Try again.';
/** Failure after a connection update attempt. */
export const ESPHOME_SAVE_RETRY_FAILED = 'Unable to save the device settings.';

/** Diagnostic code to explanatory text; unknown codes use a generic failure. */
const DIAGNOSTIC_MESSAGES: Readonly<Record<string, string>> = {
  identity_conflict: 'This address identifies a different device or a device already added.',
  entity_identity_conflict: 'An entity changed type. Review the device firmware and existing automations.',
  network_unreachable: 'The NodeHerder backend cannot reach this address. Check its network access and local network permissions.',
};
/** Generic connection failure for an absent or unknown diagnostic code. */
export const ESPHOME_CONNECTION_FAILURE =
  'Connection unavailable. Check the address, port, local network access and encryption key.';

/** Whether a port is a valid TCP port. */
export function validESPHomePort(port: number): boolean {
  return Number.isInteger(port) && port >= ESPHOME_PORT_MIN && port <= ESPHOME_PORT_MAX;
}

/** Whether the add form can be submitted; address policy remains server-side. */
export function validESPHomeInput(address: string, port: number): boolean {
  return address.trim() !== '' && validESPHomePort(port);
}

/** Builds a bounded create payload from trimmed form fields. */
export function createEnrollmentInput(name: string, address: string, port: number, key: string): ESPHomeEnrollmentInput {
  return { name: name.trim(), address: address.trim(), port, encryption_key: key.trim() };
}

/** Builds an update payload; omitted key preserves the stored credential. */
export function createEnrollmentUpdate(
  name: string,
  address: string,
  port: number,
  keyMode: ESPHomeKeyMode,
  key: string
): ESPHomeEnrollmentUpdate {
  const input: ESPHomeEnrollmentUpdate = { name: name.trim(), address: address.trim(), port };
  if (keyMode === 'replace') input.encryption_key = key.trim();
  if (keyMode === 'clear') input.encryption_key = '';
  return input;
}

/** Display label for a saved enrollment, preferring its name and then the registry. */
export function enrollmentDisplayName(item: ESPHomeEnrollment, devices: readonly Device[]): string {
  return item.name || devices.find((device) => device.id === item.id)?.friendly_name || ESPHOME_DEFAULT_DEVICE_LABEL;
}

/** Explanatory text for a classified connection diagnostic. */
export function diagnosticMessage(code?: string): string {
  return (code && DIAGNOSTIC_MESSAGES[code]) || ESPHOME_CONNECTION_FAILURE;
}

/** Initial port value for both onboarding forms. */
export const ESPHOME_FORM_DEFAULT_PORT = ESPHOME_DEFAULT_PORT;
