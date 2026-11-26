import { describe, expect, test } from '@jest/globals';
import { overviewIssues } from '@/domain/overview';
import { device, expose } from '../helpers/devices';

const NOW = Date.parse('2026-11-25T16:00:00Z');
const at = (hoursAgo: number) => new Date(NOW - hoursAgo * 60 * 60 * 1000).toISOString();

describe('overview priorities', () => {
  test('ranks offline, stale, quiet, then existing device health', () => {
    const issues = overviewIssues(
      [
        device('off', { availability: 'offline' as never, last_seen: at(2) }),
        device('stale', { last_seen: at(8) }),
        device('quiet', { last_seen: at(3) }),
        device('weak', { last_seen: at(0) }, [expose('linkquality', { data: 10 })]),
        device('fine', { last_seen: at(0) }),
      ],
      () => false,
      NOW
    );
    expect(issues.map((issue) => [issue.device.id, issue.reason])).toEqual([
      ['off', 'offline'],
      ['stale', 'stale'],
      ['quiet', 'quiet'],
      ['weak', 'weak-link'],
    ]);
  });

  test('does not duplicate a device that is both offline and quiet', () => {
    const issues = overviewIssues([device('x', { availability: 'offline' as never, last_seen: at(12) })], () => false, NOW);
    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatchObject({ device: { id: 'x' }, reason: 'offline' });
  });

  test('battery devices get the longer window and disabled devices are not quiet-flagged', () => {
    const battery = device('b', { power_source: 'battery', last_seen: at(8) });
    expect(overviewIssues([battery], () => false, NOW)).toEqual([]);
    const disabled = device('d', { last_seen: at(8) });
    expect(overviewIssues([disabled], (d) => d.id === 'd', NOW)).toEqual([
      expect.objectContaining({ reason: 'disabled' }),
    ]);
  });
});
