import { describe, expect, test } from '@jest/globals';
import { ENTITY_HELP_BY_DEVICE_CLASS, ENTITY_HELP_BY_OBJECT_ID, normalizeEntityKey } from '@/domain/entityHelp';

describe('entity help catalog data', () => {
  const entries = [...Object.values(ENTITY_HELP_BY_OBJECT_ID), ...Object.values(ENTITY_HELP_BY_DEVICE_CLASS)];

  test('every entry has non-empty plain text', () => {
    for (const entry of entries) {
      expect(entry.text.trim()).not.toBe('');
      expect(entry.text).not.toMatch(/<[^>]+>/);
    }
  });

  test('sources are https links when supplied', () => {
    for (const entry of entries.filter((item) => item.source)) {
      expect(entry.source?.url).toMatch(/^https:\/\/[^\s]+$/);
      expect(entry.source?.label.trim()).not.toBe('');
    }
  });

  test('keys are already normalized so lookup and data agree', () => {
    for (const key of Object.keys(ENTITY_HELP_BY_OBJECT_ID)) {
      expect(normalizeEntityKey(key)).toBe(key);
    }
    for (const key of Object.keys(ENTITY_HELP_BY_DEVICE_CLASS)) {
      expect(normalizeEntityKey(key)).toBe(key);
    }
  });

  test('known vendor settings and generic device classes are present', () => {
    expect(ENTITY_HELP_BY_OBJECT_ID.assume_present_timeout).toBeDefined();
    expect(ENTITY_HELP_BY_OBJECT_ID.mmwave_bluetooth).toBeDefined();
    expect(ENTITY_HELP_BY_DEVICE_CLASS.occupancy).toBeDefined();
    expect(ENTITY_HELP_BY_DEVICE_CLASS.illuminance).toBeDefined();
  });
});
