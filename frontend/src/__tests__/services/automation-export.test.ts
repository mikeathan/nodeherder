import { describe, expect, test } from '@jest/globals';
import { serializeAutomations } from '@/services/automation-export';
import type { Automation } from '@/types/automation.type';

const automation = { id: 'a1', friendlyname: 'Hallway', enabled: true, triggers: [] } as unknown as Automation;

describe('automation export', () => {
  test('wraps a versioned, timestamped envelope', () => {
    const parsed = JSON.parse(serializeAutomations([automation], new Date('2026-11-25T16:00:00Z')));
    expect(parsed).toMatchObject({ version: 1, exportedAt: '2026-11-25T16:00:00.000Z' });
    expect(parsed.automations).toHaveLength(1);
    expect(parsed.automations[0].id).toBe('a1');
  });
});
