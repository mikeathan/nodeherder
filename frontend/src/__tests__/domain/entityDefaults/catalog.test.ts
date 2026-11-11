import { describe, expect, test } from '@jest/globals';
import { ENTITY_DEFAULTS_BY_OBJECT_ID } from '@/domain/entityDefaults';
import { normalizeEntityKey } from '@/domain/entityKey';

describe('entity defaults catalog data', () => {
  test('keys are already normalized so lookup and data agree', () => {
    for (const key of Object.keys(ENTITY_DEFAULTS_BY_OBJECT_ID)) {
      expect(normalizeEntityKey(key)).toBe(key);
    }
  });

  test('every default is a scalar and sources are https links', () => {
    for (const entry of Object.values(ENTITY_DEFAULTS_BY_OBJECT_ID)) {
      expect(['number', 'string', 'boolean']).toContain(typeof entry.value);
      if (typeof entry.value === 'number') expect(Number.isFinite(entry.value)).toBe(true);
      if (typeof entry.value === 'string') expect(entry.value.trim()).not.toBe('');
      if (entry.source) expect(entry.source.url).toMatch(/^https:\/\/[^\s]+$/);
    }
  });

  test('stateless and disruptive entities are excluded', () => {
    expect(ENTITY_DEFAULTS_BY_OBJECT_ID.esp_reboot).toBeUndefined();
    expect(ENTITY_DEFAULTS_BY_OBJECT_ID.factory_reset_mmwave_sensor).toBeUndefined();
    expect(ENTITY_DEFAULTS_BY_OBJECT_ID.everything_presence_lite_firmware).toBeUndefined();
    expect(ENTITY_DEFAULTS_BY_OBJECT_ID.mmwave_bluetooth).toBeUndefined();
  });
});
