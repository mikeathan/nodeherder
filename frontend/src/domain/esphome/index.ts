/*
 * Pure ESPHome onboarding helpers: input validation, payload construction and
 * user-facing labels. Dialogs keep only presentation and lifecycle concerns.
 */
import type { Device } from '@/types/device';
import {
  ESPHOME_DEFAULT_DEVICE_LABEL,
  ESPHOME_MESSAGES,
  ESPHOME_PORT_MAX,
  ESPHOME_PORT_MIN,
  ESPHOME_STATUS_LABELS,
  type ESPHomeEnrollment,
  type ESPHomeEnrollmentInput,
  type ESPHomeEnrollmentUpdate,
} from '@/contracts/esphome';

/** How the edit form treats the stored encryption key. */
export type ESPHomeKeyMode = 'keep' | 'replace' | 'clear';

/** Diagnostic code to explanatory text; unknown codes use a generic failure. */
const DIAGNOSTIC_MESSAGES: Readonly<Record<string, string>> = {
  identity_conflict: 'This address identifies a different device or a device already added.',
  entity_identity_conflict: 'An entity changed type. Review the device firmware and existing automations.',
  network_unreachable: 'The NodeHerder backend cannot reach this address. Check its network access and local network permissions.',
};
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
  return (code && DIAGNOSTIC_MESSAGES[code]) || ESPHOME_MESSAGES.connectionFailure;
}

/** Display label for a reported connection status; unknown values are stated, not echoed raw. */
export function connectionStatusLabel(status?: string): string {
  return ESPHOME_STATUS_LABELS[status ?? ''] ?? ESPHOME_STATUS_LABELS.unknown;
}
