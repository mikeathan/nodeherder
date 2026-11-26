import { describe, expect, test } from '@jest/globals';
import { isAmbientNight, parsePanelAmbient } from '@/domain/panel-ambient';

describe('panel ambient light', () => {
  test('parses a bounded preference and rejects malformed data', () => {
    expect(parsePanelAmbient({ deviceId: 'd', expose: 'illuminance', threshold: 20 })).toEqual({ deviceId: 'd', expose: 'illuminance', threshold: 20 });
    expect(parsePanelAmbient(null)).toBeNull();
    expect(parsePanelAmbient({ deviceId: 'd', expose: 'illuminance', threshold: -1 })).toBeNull();
    expect(parsePanelAmbient({ deviceId: '', expose: 'illuminance', threshold: 20 })).toBeNull();
  });

  test('night starts only below the threshold with a finite reading', () => {
    expect(isAmbientNight(5, 20)).toBe(true);
    expect(isAmbientNight(20, 20)).toBe(false);
    expect(isAmbientNight(undefined, 20)).toBe(false);
    expect(isAmbientNight('dark', 20)).toBe(false);
  });
});
