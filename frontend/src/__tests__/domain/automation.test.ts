import { cloneAutomation, describeTrigger, issuesUnder, newAutomation, sameAutomation, validateAutomation } from '@/domain/automation';
import { Automation, AutomationTrigger } from '@/types/automation.type';
import { device, expose } from '../helpers/devices';

const lamp = device('lamp', { friendly_name: 'Desk lamp' }, [
  expose('state', { type: 'binary' as never, access_mode: 'readwrite' as never, values: { on: 'ON', off: 'OFF' } }),
  expose('brightness', { access_mode: 'readwrite' as never, attributes: { min: 0, max: 254 } }),
]);
const sensor = device('pir', { friendly_name: 'Hall sensor' }, [expose('occupancy', { type: 'binary' as never, values: { on: true, off: false } })]);
const offline = device('plug', { friendly_name: 'Plug', availability: 'offline' as never }, [expose('state')]);
const lookup = (id: string) => [lamp, sensor, offline].find((d) => d.id === id);

const trigger = (partial: Partial<AutomationTrigger> = {}): AutomationTrigger => ({
  name: 'occupancy',
  type: 'deviceTrigger',
  conditions: [{ type: 'expose', name: 'occupancy', value: true, equality: '=' }],
  actions: [{ id: 'lamp', type: 'trigger', exposes: [{ name: 'state', data: 'ON' }], publishMode: 'batch' } as never],
  ...partial,
});
const automation = (triggers: AutomationTrigger[]): Automation => ({
  id: 'pir', friendlyname: 'Hall sensor', type: 'device', description: '', enabled: true, schedules: [], triggers,
});

describe('describeTrigger', () => {
  it('summarises a device trigger as a sentence', () => {
    const a = automation([trigger()]);
    expect(describeTrigger(a, a.triggers[0], lookup)).toBe('When Hall sensor occupancy changes, if occupancy = Detected, then set Desk lamp state to On.');
  });

  it('describes time conditions and delays', () => {
    const t = trigger({
      conditions: [{ type: 'time', timeRange: { startAt: '22:00', endAt: '06:00' } }],
      actions: [{ id: 'lamp', type: 'trigger', exposes: [{ name: 'brightness', data: 127 }], publishMode: 'batch', delay: { value: 5, unit: 'seconds' } } as never],
    });
    expect(describeTrigger(automation([t]), t, lookup)).toBe('When Hall sensor occupancy changes, if the time is between 22:00 and 06:00 (overnight), then set Desk lamp brightness to 50 % after 5 seconds.');
  });

  it('handles empty states', () => {
    const a = automation([]);
    expect(describeTrigger(a, undefined, lookup)).toBe('No triggers yet.');
    expect(describeTrigger(a, trigger({ conditions: [], actions: [] }), lookup)).toMatch(/, then …\.$/);
  });
});

describe('validateAutomation', () => {
  it('accepts a complete automation', () => {
    expect(validateAutomation(automation([trigger()]), lookup)).toEqual({ errors: [], warnings: [] });
  });

  it('reports field errors with dot paths', () => {
    const t = trigger({
      name: '',
      conditions: [{ type: 'time', timeRange: { startAt: '25:00', endAt: '' } }, { type: 'expose', name: 'occupancy', value: '', equality: '=' }],
      actions: [
        { id: '', type: 'trigger', exposes: [] } as never,
        { id: 'lamp', type: 'trigger', exposes: [{ name: 'state', data: null }], delay: { value: -1, unit: 'seconds' } } as never,
        { id: 'lamp', type: 'step', property: '', steps: [], data: 'x' } as never,
        { id: 'lamp', type: 'preset', property: '' } as never,
      ],
    });
    const paths = validateAutomation(automation([t]), lookup).errors.map((e) => e.path);
    expect(paths).toEqual([
      'triggers.0.name',
      'triggers.0.conditions.0.timeRange.startAt',
      'triggers.0.conditions.0.timeRange.endAt',
      'triggers.0.conditions.1.value',
      'triggers.0.actions.0.id',
      'triggers.0.actions.1.exposes.0.data',
      'triggers.0.actions.1.delay.value',
      'triggers.0.actions.2.property',
      'triggers.0.actions.2.data',
      'triggers.0.actions.2.steps',
      'triggers.0.actions.3.property',
    ]);
  });

  it('warns about no triggers, missing conditions, offline targets and loops', () => {
    expect(validateAutomation(automation([]), lookup).warnings.map((w) => w.path)).toEqual(['triggers']);
    const t = trigger({
      conditions: [],
      actions: [
        { id: 'plug', type: 'trigger', exposes: [{ name: 'state', data: 'ON' }] } as never,
        { id: 'pir', type: 'trigger', exposes: [{ name: 'occupancy', data: false }] } as never,
      ],
    });
    const { errors, warnings } = validateAutomation(automation([t]), lookup);
    expect(errors).toEqual([]);
    expect(warnings.map((w) => w.path)).toEqual(['triggers.0', 'triggers.0.actions.0', 'triggers.0.actions.1']);
  });

  it('treats a trigger without a type as a device trigger (legacy payloads)', () => {
    const t = { ...trigger({ conditions: [] }), type: undefined } as unknown as AutomationTrigger;
    expect(validateAutomation(automation([t]), lookup).warnings.map((w) => w.path)).toContain('triggers.0');
  });

  it('flags actions targeting removed devices', () => {
    const t = trigger({ actions: [{ id: 'gone', type: 'preset', property: 'mode' } as never] });
    expect(validateAutomation(automation([t]), lookup).errors).toEqual([{ path: 'triggers.0.actions.0.id', message: 'This device no longer exists.' }]);
  });
});

describe('issuesUnder', () => {
  it('matches the path and its children only', () => {
    const issues = [{ path: 'triggers.1', message: 'a' }, { path: 'triggers.1.name', message: 'b' }, { path: 'triggers.10', message: 'c' }];
    expect(issuesUnder(issues, 'triggers.1').map((i) => i.message)).toEqual(['a', 'b']);
  });
});

describe('new automations', () => {
  it('starts with one empty device trigger and is disabled', () => {
    const a = newAutomation({ id: 'pir', friendly_name: 'Hall sensor' });
    expect(a).toEqual({ id: 'pir', friendlyname: 'Hall sensor', type: 'device', description: '', enabled: false, schedules: [], triggers: [{ name: '', type: 'deviceTrigger', conditions: [], actions: [] }] });
  });

  it('clones without sharing state and compares by payload', () => {
    const a = automation([trigger()]);
    const b = cloneAutomation(a);
    expect(sameAutomation(a, b)).toBe(true);
    b.triggers[0].conditions = [];
    expect(sameAutomation(a, b)).toBe(false);
    expect(a.triggers[0].conditions).toHaveLength(1);
  });
});

describe('step actions', () => {
  it('describe the backend formula, innermost step last', () => {
    const t = trigger({
      actions: [{ id: 'lamp', type: 'step', property: 'brightness', data: 0.5, steps: [{ id: 'lamp', property: 'brightness', operator: '+' }, { id: 'pir', property: 'action_time', operator: '*' }] } as never],
    });
    expect(describeTrigger(automation([t]), t, lookup)).toBe('When Hall sensor occupancy changes, if occupancy = Detected, then set Desk lamp brightness to brightness + (action time × 0.5).');
  });
});
