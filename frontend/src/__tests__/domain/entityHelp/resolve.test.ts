import { describe, expect, test } from '@jest/globals';
import {
  catalogEntityHelp,
  normalizeEntityKey,
  providerEntityHelp,
  resolveEntityHelp,
} from '@/domain/entityHelp';
import { expose } from '../../helpers/devices';

function esphomeEntity(attributes: Record<string, unknown>, description = 'Assume Present Timeout') {
  return expose('00000000-0e4537ac', {
    description,
    category: 'config' as never,
    access_mode: 'readwrite' as never,
    attributes: { label: 'Assume Present Timeout', entity_kind: 'number', ...attributes } as never,
  });
}

describe('entity key normalization', () => {
  test.each([
    ['Assume Present Timeout', 'assume_present_timeout'],
    ['mmWave Bluetooth', 'mmwave_bluetooth'],
    ['ESP32 LED', 'esp32_led'],
    ['Factory Reset mmWave Sensor', 'factory_reset_mmwave_sensor'],
    ['  Extra   entities update  ', 'extra_entities_update'],
    ['Exit Threshold Pct', 'exit_threshold_pct'],
  ])('maps %p to %p', (input, expected) => {
    expect(normalizeEntityKey(input)).toBe(expected);
  });

  test('returns an empty key for non-strings', () => {
    expect(normalizeEntityKey(undefined)).toBe('');
    expect(normalizeEntityKey(42)).toBe('');
  });
});

describe('entity help resolution', () => {
  test('matches an ESPHome object_id and keeps its documentation source', () => {
    const help = resolveEntityHelp(esphomeEntity({ object_id: 'assume_present_timeout' }));
    expect(help?.kind).toBe('catalog');
    expect(help?.text).toContain('hold occupancy');
    expect(help?.source?.url).toContain('docs.everythingsmart.io');
  });

  test('falls back to the normalized label when object_id is absent', () => {
    expect(resolveEntityHelp(esphomeEntity({}))?.text).toContain('hold occupancy');
  });

  test('prefers a distinct provider description over catalog data', () => {
    const help = resolveEntityHelp(
      expose('occupancy', { description: 'Provider supplied explanation', attributes: { device_class: 'occupancy' } as never })
    );
    expect(help).toEqual({ text: 'Provider supplied explanation', kind: 'provider' });
  });

  test('ignores a provider description that repeats the label', () => {
    expect(providerEntityHelp(esphomeEntity({ object_id: 'assume_present_timeout' }))).toBeNull();
    expect(resolveEntityHelp(esphomeEntity({ object_id: 'assume_present_timeout' }))?.kind).toBe('catalog');
  });

  test('matches generic device classes for any provider', () => {
    const help = resolveEntityHelp(expose('00000000-12345678', { attributes: { device_class: 'occupancy' } as never }));
    expect(help?.kind).toBe('catalog');
    expect(help?.text).toContain('presence');
  });

  test('returns nothing for an unknown entity instead of inventing text', () => {
    const unknown = expose('mystery_setting', {
      description: 'Mystery Setting',
      attributes: { label: 'Mystery Setting' } as never,
    });
    expect(catalogEntityHelp(unknown)).toBeNull();
    expect(resolveEntityHelp(unknown)).toBeNull();
  });

  test('object_id coverage does not depend on the display label', () => {
    const renamed = expose('00000000-0e4537ac', {
      description: 'Occupancy hold',
      attributes: { label: 'Occupancy hold', entity_kind: 'number', object_id: 'assume_present_timeout' } as never,
    });
    expect(resolveEntityHelp(renamed)?.text).toContain('hold occupancy');
  });
});
