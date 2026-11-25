/*
 * Overview priorities: one ranked list of things a person can act on, combining existing
 * device health with reporting freshness. Freshness stays display-only here; this module
 * never changes `deviceStatus` or command gating.
 */
import type { Device } from '@/types/device';
import { compareByName } from './devices';
import { freshnessPolicy, reportFreshness } from './devices/freshness';
import { needsAttention, type AttentionReason } from './network';

/** Reasons shown in the calm Overview, ranked by consequence. */
export type OverviewReason = AttentionReason | 'stale' | 'quiet';

export type OverviewIssue = { device: Device; reason: OverviewReason };

const RANK: Record<OverviewReason, number> = {
  offline: 0,
  stale: 1,
  quiet: 2,
  'low-battery': 3,
  'weak-link': 4,
  disabled: 5,
};

/** One row per device, most consequential first, then name. */
export function overviewIssues(devices: Device[], isDisabled: (device: Device) => boolean, now: number): OverviewIssue[] {
  const issues: OverviewIssue[] = needsAttention(devices, isDisabled).map((item) => ({ device: item.device, reason: item.reason }));
  const claimed = new Set(issues.map((issue) => issue.device.id));
  for (const device of devices) {
    if (claimed.has(device.id)) continue;
    if (isDisabled(device)) continue;
    const freshness = reportFreshness(device.last_seen, now, freshnessPolicy(device));
    if (freshness === 'quiet' || freshness === 'stale') {
      issues.push({ device, reason: freshness });
    }
  }
  return issues.sort((left, right) => RANK[left.reason] - RANK[right.reason] || compareByName(left.device, right.device));
}

/** Human label for one reason; UI owns tone/icon mapping. */
export const OVERVIEW_REASON_LABEL: Readonly<Record<OverviewReason, string>> = {
  offline: 'Offline',
  stale: 'Not reporting',
  quiet: 'Quiet',
  'low-battery': 'Low battery',
  'weak-link': 'Weak signal',
  disabled: 'Disabled',
};
