import { describe, expect, test } from '@jest/globals';
import {
  BATTERY_FRESHNESS,
  freshnessPolicy,
  MAINS_FRESHNESS,
  reportAgeMs,
  reportFreshness,
} from '@/domain/devices/freshness';
import { device } from '../../helpers/devices';

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const NOW = Date.parse('2026-11-25T16:00:00Z');
const iso = (offsetMs: number) => new Date(NOW - offsetMs).toISOString();

describe('reporting freshness', () => {
  test('mains devices become quiet then stale, and never change availability', () => {
    expect(reportFreshness(iso(18 * 1000), NOW, MAINS_FRESHNESS)).toBe('fresh');
    expect(reportFreshness(iso(2 * HOUR), NOW, MAINS_FRESHNESS)).toBe('quiet');
    expect(reportFreshness(iso(6 * HOUR), NOW, MAINS_FRESHNESS)).toBe('stale');
  });

  test('battery devices use the longer window', () => {
    expect(reportFreshness(iso(4 * HOUR), NOW, BATTERY_FRESHNESS)).toBe('fresh');
    expect(reportFreshness(iso(24 * HOUR), NOW, BATTERY_FRESHNESS)).toBe('quiet');
    expect(reportFreshness(iso(48 * HOUR), NOW, BATTERY_FRESHNESS)).toBe('stale');
  });

  test('policy follows the reported power source', () => {
    expect(freshnessPolicy(device('mains'))).toEqual(MAINS_FRESHNESS);
    expect(freshnessPolicy(device('battery', { power_source: 'battery' }))).toEqual(BATTERY_FRESHNESS);
    expect(freshnessPolicy(undefined)).toEqual(MAINS_FRESHNESS);
  });

  test('missing, malformed and future timestamps are handled without throwing', () => {
    expect(reportFreshness('', NOW)).toBe('unknown');
    expect(reportFreshness('not-a-date', NOW)).toBe('unknown');
    expect(reportAgeMs('not-a-date', NOW)).toBeNull();
    expect(reportAgeMs(iso(-5 * MINUTE), NOW)).toBe(0);
    expect(reportFreshness(iso(-5 * MINUTE), NOW)).toBe('fresh');
  });

  test('age is the difference between now and the report', () => {
    expect(reportAgeMs(iso(90 * MINUTE), NOW)).toBe(90 * MINUTE);
  });
});
