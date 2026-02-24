import {
  bucketBinaryEvents,
  BINARY_NOISE_THRESHOLD,
  renderHeatmapTooltip,
  getBinaryStats,
  isBinaryOn,
} from '../../utils/chart.utils';
import type { BinaryDataPoint } from '../../types/metrics.type';
import { describe, expect, it } from '@jest/globals';

describe('bucketBinaryEvents', () => {
  const makePoint = (timestamp: number, value: string): BinaryDataPoint => ({
    timestamp,
    value,
  });

  const HOUR = 60 * 60_000;

  it('returns empty array for empty data', () => {
    expect(bucketBinaryEvents([], 0, HOUR)).toEqual([]);
  });

  it('returns empty array when from >= to', () => {
    const data = [makePoint(1000, 'true')];
    expect(bucketBinaryEvents(data, HOUR, 0)).toEqual([]);
    expect(bucketBinaryEvents(data, HOUR, HOUR)).toEqual([]);
  });

  it('creates correct number of buckets for 1 hour with 15min intervals', () => {
    const from = 0;
    const to = HOUR;
    const data = [makePoint(100, 'true'), makePoint(HOUR - 100, 'false')];
    const buckets = bucketBinaryEvents(data, from, to, 15);

    expect(buckets).toHaveLength(4); // 60min / 15min = 4
    expect(buckets[0].start).toBe(from);
    expect(buckets[3].end).toBe(to);
  });

  it('counts transitions correctly', () => {
    const from = 0;
    const to = HOUR;
    // All transitions happen in the second bucket (15-30 min)
    const data = [makePoint(16 * 60_000, 'true'), makePoint(20 * 60_000, 'false'), makePoint(25 * 60_000, 'true')];
    const buckets = bucketBinaryEvents(data, from, to, 15);

    // Bucket 0 (0-15min): no transitions started here
    expect(buckets[0].count).toBe(0);
    // Bucket 1 (15-30min): 3 transitions started here
    expect(buckets[1].count).toBe(3);
    // Bucket 2-3: no transitions
    expect(buckets[2].count).toBe(0);
    expect(buckets[3].count).toBe(0);
  });

  it('normalizes intensity to 0..1', () => {
    const from = 0;
    const to = HOUR;
    const data = [makePoint(16 * 60_000, 'true'), makePoint(20 * 60_000, 'false'), makePoint(25 * 60_000, 'true')];
    const buckets = bucketBinaryEvents(data, from, to, 15);

    // Max count is 3 (in bucket index 1)
    expect(buckets[1].intensity).toBe(1);
    expect(buckets[0].intensity).toBe(0);
  });

  it('calculates active duration within buckets', () => {
    const from = 0;
    const to = 30 * 60_000; // 30 minutes
    // Active from 5min to 25min
    const data = [makePoint(5 * 60_000, 'true'), makePoint(25 * 60_000, 'false')];
    const buckets = bucketBinaryEvents(data, from, to, 15);

    // Bucket 0 (0-15min): normalizeBinaryEvents creates a leading segment
    // from 0→5min with the first event's value ('true'), plus 5→15min active = 15min total
    expect(buckets[0].activeMs).toBe(15 * 60_000);
    // Bucket 1 (15-30min): active from 15min to 25min = 10min
    expect(buckets[1].activeMs).toBe(10 * 60_000);
  });
});

describe('BINARY_NOISE_THRESHOLD', () => {
  it('is a positive number', () => {
    expect(BINARY_NOISE_THRESHOLD).toBeGreaterThan(0);
    expect(BINARY_NOISE_THRESHOLD).toBe(50);
  });
});

describe('renderHeatmapTooltip', () => {
  it('returns HTML containing expected values', () => {
    const bucket = {
      start: 1700000000000,
      end: 1700000900000, // 15 min later
      count: 14,
      activeMs: 600000, // 10 min
      intensity: 0.7,
    };
    const html = renderHeatmapTooltip(bucket, 'presence');

    expect(html).toContain('presence');
    expect(html).toContain('14');
    expect(html).toContain('Triggers');
  });
});

describe('getBinaryStats', () => {
  it('returns zero stats for empty data', () => {
    const stats = getBinaryStats([]);
    expect(stats.onCount).toBe(0);
    expect(stats.offCount).toBe(0);
    expect(stats.onPercentage).toBe(0);
  });

  it('calculates correct on/off counts', () => {
    const now = Date.now();
    const data: BinaryDataPoint[] = [
      { timestamp: now - 3000, value: 'true' },
      { timestamp: now - 2000, value: 'false' },
      { timestamp: now - 1000, value: 'true' },
    ];
    const stats = getBinaryStats(data, now);
    expect(stats.onCount).toBe(2);
    expect(stats.offCount).toBe(1);
  });
});

describe('isBinaryOn', () => {
  it('handles boolean true', () => expect(isBinaryOn(true)).toBe(true));
  it('handles boolean false', () => expect(isBinaryOn(false)).toBe(false));
  it('handles string "true"', () => expect(isBinaryOn('true')).toBe(true));
  it('handles string "false"', () => expect(isBinaryOn('false')).toBe(false));
  it('handles string "on"', () => expect(isBinaryOn('on')).toBe(true));
  it('handles string "1"', () => expect(isBinaryOn('1')).toBe(true));
  it('handles string "off"', () => expect(isBinaryOn('off')).toBe(false));
});
