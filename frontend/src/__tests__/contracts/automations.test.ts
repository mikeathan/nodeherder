import { automationPayload, createActionFromType } from '@/contracts/automations';
import { Automation, AutomationTriggerAction } from '@/types/automation.type';

describe('automationPayload', () => {
  const recipe = (delay?: AutomationTriggerAction['delay']): Automation => ({
    id: 'dial', friendlyname: 'Test dial', type: 'device', description: '', enabled: true, schedules: [],
    triggers: [{ name: 'action', type: 'deviceTrigger', conditions: [], actions: [{
      ...createActionFromType('trigger'), id: 'light', type: 'trigger',
      exposes: [{ name: 'state', data: 'TOGGLE' }, { name: 'brightness', data: 0 }], delay,
    }] }],
  });

  it('omits a zero wait, preserving native commands and the local draft', () => {
    const a = recipe({ unit: 'minutes', value: 0 });
    const before = JSON.stringify(a);
    const payload = automationPayload(a);
    expect(payload.triggers[0].actions[0]).not.toHaveProperty('delay');
    expect((payload.triggers[0].actions[0] as AutomationTriggerAction).exposes).toEqual([
      { name: 'state', data: 'TOGGLE' }, { name: 'brightness', data: 0 },
    ]);
    expect(JSON.stringify(a)).toBe(before);
  });

  it.each([undefined, { unit: 'seconds' as const, value: 5 }, { unit: 'hours' as const, value: 1 }])(
    'preserves valid existing payload bytes for delay %p', (delay) => {
      const a = recipe(delay);
      expect(JSON.stringify(automationPayload(a))).toBe(JSON.stringify(a));
    }
  );
});
