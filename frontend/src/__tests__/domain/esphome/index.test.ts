import { describe, expect, test } from '@jest/globals';
import {
  createEnrollmentInput,
  createEnrollmentUpdate,
  diagnosticMessage,
  connectionStatusLabel,
  enrollmentDisplayName,
  validESPHomeInput,
  validESPHomePort,
} from '@/domain/esphome';
import { ESPHOME_DEFAULT_DEVICE_LABEL, ESPHOME_MESSAGES } from '@/contracts/esphome';
import { device } from '../../helpers/devices';

describe('ESPHome onboarding helpers', () => {
  test('validates the port range and required address', () => {
    expect(validESPHomePort(1)).toBe(true);
    expect(validESPHomePort(65535)).toBe(true);
    expect(validESPHomePort(0)).toBe(false);
    expect(validESPHomePort(65536)).toBe(false);
    expect(validESPHomePort(6053.5)).toBe(false);
    expect(validESPHomeInput('  ', 6053)).toBe(false);
    expect(validESPHomeInput('192.168.1.2', 6053)).toBe(true);
  });

  test('trims create input without inventing a key', () => {
    expect(createEnrollmentInput(' Radar ', ' 192.168.1.2 ', 6053, ' ')).toEqual({
      name: 'Radar',
      address: '192.168.1.2',
      port: 6053,
      encryption_key: '',
    });
  });

  test('update input keeps, replaces or clears the stored key', () => {
    expect(createEnrollmentUpdate('Radar', '192.168.1.2', 6053, 'keep', 'ignored')).toEqual({
      name: 'Radar',
      address: '192.168.1.2',
      port: 6053,
    });
    expect(createEnrollmentUpdate('Radar', '192.168.1.2', 6053, 'replace', ' key ')).toEqual({
      name: 'Radar',
      address: '192.168.1.2',
      port: 6053,
      encryption_key: 'key',
    });
    expect(createEnrollmentUpdate('Radar', '192.168.1.2', 6053, 'clear', 'ignore').encryption_key).toBe('');
  });

  test('reported connection status is labelled, never echoed raw', () => {
    // The hub sends "connecting"/"online"/"offline"; anything else is not a status this UI knows.
    expect(connectionStatusLabel('connecting')).toBe('Connecting');
    expect(connectionStatusLabel('online')).toBe('Online');
    expect(connectionStatusLabel('offline')).toBe('Offline');
    expect(connectionStatusLabel(undefined)).toBe('Not reported');
    expect(connectionStatusLabel('something-new')).toBe('Not reported');
  });

  test('display name and diagnostics have safe fallbacks', () => {
    expect(enrollmentDisplayName({ id: 'id', address: '192.168.1.2', port: 6053 }, [device('id', { friendly_name: 'Radar' })])).toBe('Radar');
    expect(enrollmentDisplayName({ id: 'other', address: '192.168.1.2', port: 6053 }, [])).toBe(ESPHOME_DEFAULT_DEVICE_LABEL);
    expect(diagnosticMessage('identity_conflict')).toContain('different device');
    expect(diagnosticMessage('unknown')).toBe(ESPHOME_MESSAGES.connectionFailure);
  });
});
