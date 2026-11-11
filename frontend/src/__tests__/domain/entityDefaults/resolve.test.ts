import { describe, expect, test } from '@jest/globals';
import { entityDefaultTargets, resetDefaultsMessage, resolveEntityDefault } from '@/domain/entityDefaults';
import { expose } from '../../helpers/devices';

function numberEntity(objectID: string, label: string, attributes: Record<string, unknown> = {}) {
  return expose(objectID, {
    description: label,
    type: 'numeric' as never,
    category: 'config' as never,
    access_mode: 'readwrite' as never,
    attributes: { label, object_id: objectID, entity_kind: 'number', ...attributes } as never,
  });
}

function switchEntity(objectID: string, label: string) {
  return expose(objectID, {
    description: label,
    type: 'binary' as never,
    category: 'config' as never,
    access_mode: 'readwrite' as never,
    attributes: { label, object_id: objectID, entity_kind: 'switch' } as never,
    values: { on: true, off: false } as never,
  });
}

function selectEntity(objectID: string, label: string, options: string[]) {
  return expose(objectID, {
    description: label,
    type: 'enum' as never,
    category: 'config' as never,
    access_mode: 'readwrite' as never,
    attributes: { label, object_id: objectID, entity_kind: 'select' } as never,
    values: Object.fromEntries(options.map((option, index) => [String(index), option])) as never,
  });
}

describe('entity default resolution', () => {
  test('resolves a documented number default with its source', () => {
    const entity = numberEntity('assume_present_timeout', 'Assume Present Timeout', { min: 0, max: 43200, step: 1 });
    const target = resolveEntityDefault(entity);
    expect(target?.value).toBe(1800);
    expect(target?.source?.url).toContain('docs.everythingsmart.io');
  });

  test('falls back to the normalized label when object_id is absent', () => {
    const entity = numberEntity('', 'Assume Present Timeout', { min: 0, max: 43200, step: 1 });
    expect(resolveEntityDefault(entity)?.value).toBe(1800);
  });

  test('resolves boolean switch and light defaults', () => {
    expect(resolveEntityDefault(switchEntity('entry_exit_enabled', 'Entry Exit Enabled'))?.value).toBe(false);
    expect(resolveEntityDefault(switchEntity('esp32_led', 'ESP32 LED'))?.value).toBe(false);
  });

  test('resolves a select default only when the option still exists', () => {
    const present = selectEntity('tracking_behaviour', 'Tracking Behaviour', [
      'None',
      'Above + Speed and Resolution',
    ]);
    expect(resolveEntityDefault(present)?.value).toBe('Above + Speed and Resolution');
    expect(resolveEntityDefault(selectEntity('tracking_behaviour', 'Tracking Behaviour', ['None']))).toBeNull();
  });

  test('skips a default that no longer fits the control type, range or step', () => {
    expect(resolveEntityDefault(selectEntity('assume_present_timeout', 'Assume Present Timeout', ['one']))).toBeNull();
    expect(
      resolveEntityDefault(numberEntity('exit_threshold_pct', 'Exit Threshold Pct', { min: 0, max: 5, step: 1 }))
    ).toBeNull();
    expect(
      resolveEntityDefault(numberEntity('exit_threshold_pct', 'Exit Threshold Pct', { min: 0, max: 100, step: 3 }))
    ).toBeNull();
  });

  test('skips unknown, button and unsupported entities', () => {
    expect(resolveEntityDefault(numberEntity('mystery_setting', 'Mystery Setting'))).toBeNull();
    const button = expose('factory_reset_mmwave_sensor', {
      type: 'button' as never,
      access_mode: 'write' as never,
      category: 'config' as never,
      attributes: {
        label: 'Factory Reset mmWave Sensor',
        object_id: 'factory_reset_mmwave_sensor',
        entity_kind: 'button',
      } as never,
    });
    expect(resolveEntityDefault(button)).toBeNull();
  });

  test('filters a whole entity list and builds confirmation copy', () => {
    const targets = entityDefaultTargets([
      numberEntity('assume_present_timeout', 'Assume Present Timeout', { min: 0, max: 43200, step: 1 }),
      numberEntity('mystery_setting', 'Mystery Setting'),
      switchEntity('esp32_led', 'ESP32 LED'),
    ]);
    expect(targets.map((target) => target.expose.name)).toEqual(['assume_present_timeout', 'esp32_led']);
    const message = resetDefaultsMessage(targets);
    expect(message).toContain('Reset 2 documented settings');
    expect(message).toContain('Assume Present Timeout');
    expect(message).toContain('only after the device confirms');
    expect(message).not.toContain('reset succeeded');
  });
});
