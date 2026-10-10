import { describe, expect, test } from '@jest/globals';
import { capabilityOf, controlOf, exposeKind, exposeLabel, exposePresets, exposeRange, formatExposeValue, isBinaryOn, stateExposeOf } from '@/domain/exposes';
import { device, expose, hubDevices } from '../helpers/devices';

describe('expose capabilities', () => {
  test('writable binary is a switch, read-only binary a sensor', () => {
    expect(capabilityOf(expose('state', { type: 'binary' as never, access_mode: 'readwrite' as never }))).toBe('switch');
    expect(capabilityOf(expose('contact', { type: 'binary' as never }))).toBe('binary-sensor');
  });
  test('writable numeric with a max is a slider; with values only a preset', () => {
    expect(capabilityOf(expose('brightness', { access_mode: 'readwrite' as never, attributes: { min: 0, max: 254 } }))).toBe('slider');
    expect(capabilityOf(expose('effect', { type: 'enum' as never, access_mode: 'write' as never, values: { blink: 'blink' } }))).toBe('preset');
  });
  test('config, diagnostic, action and reading', () => {
    expect(capabilityOf(expose('child_lock', { category: 'config' as never }))).toBe('config');
    expect(capabilityOf(expose('linkquality', { category: 'diagnostic' as never }))).toBe('diagnostic');
    expect(capabilityOf(expose('action', { type: 'enum' as never }))).toBe('event');
    expect(capabilityOf(expose('temperature'))).toBe('reading');
  });
  test('every expose of every fixture device resolves without throwing', () => {
    for (const d of hubDevices()) for (const e of Object.values(d.exposes)) expect(typeof capabilityOf(e)).toBe('string');
  });
});

describe('expose presentation', () => {
  test('kinds and friendly labels with fallback', () => {
    expect(exposeKind('temperature')).toBe('temp');
    expect(exposeKind('energy_month')).toBe('energy');
    expect(exposeKind('made_up')).toBe('misc');
    expect(exposeLabel('co2')).toBe('CO₂');
    expect(exposeLabel('linkquality')).toBe('Link quality');
    expect(exposeLabel('target_distance')).toBe('Target distance');
    expect(exposeLabel('some_new_thing_here')).toBe('Some new thing here');
  });
  test('brightness is a percentage of max, never "254%"', () => {
    const b = expose('brightness', { data: 254, attributes: { min: 0, max: 254 } });
    expect(formatExposeValue(b)).toBe('100 %');
    expect(formatExposeValue(b, 127)).toBe('50 %');
  });
  test('units, colour temperature, words and missing data', () => {
    expect(formatExposeValue(expose('temperature', { data: 19.84, unit: '°C' }))).toBe('19.8 °C');
    expect(formatExposeValue(expose('voltage', { data: 3.105, unit: 'V' }))).toBe('3.1 V');
    expect(formatExposeValue(expose('linkquality', { data: 91, unit: 'lqi' }))).toBe('91');
    expect(formatExposeValue(expose('color_temp', { data: 370 }))).toBe('2703 K');
    expect(formatExposeValue(expose('state', { type: 'binary' as never, data: 'ON', values: { on: 'ON', off: 'OFF' } }))).toBe('On');
    // Zigbee2MQTT: contact on=false means open, off=true means closed
    const contact = { type: 'binary' as never, values: { on: false, off: true } };
    expect(formatExposeValue(expose('contact', { ...contact, data: false }))).toBe('Open');
    expect(formatExposeValue(expose('contact', { ...contact, data: true }))).toBe('Closed');
    expect(formatExposeValue(expose('presence', { type: 'binary' as never, data: true }))).toBe('Detected');
    expect(formatExposeValue(expose('action', { type: 'enum' as never, data: 'brightness_step_up' }))).toBe('Brightness step up');
    expect(formatExposeValue(expose('temperature', { data: null }))).toBe('—');
    expect(formatExposeValue(expose('power', { data: 0, unit: 'W' }))).toBe('0 W');
  });
  test('binary on detection uses values.on when present', () => {
    const s = expose('state', { type: 'binary' as never, values: { on: 'ON', off: 'OFF' } });
    expect(isBinaryOn(s, 'ON')).toBe(true);
    expect(isBinaryOn(s, 'OFF')).toBe(false);
  });
  test('state expose lookup', () => {
    const st = expose('state', { type: 'binary' as never, access_mode: 'readwrite' as never, values: { on: 'ON', off: 'OFF' } });
    expect(stateExposeOf(device('a', {}, [expose('brightness'), st]))?.name).toBe('state');
    expect(stateExposeOf(device('b', {}, [expose('temperature')]))).toBeUndefined();
  });
});

describe('expose ranges and presets', () => {
  test('reads hub min/max and Zigbee2MQTT value_min/value_max', () => {
    expect(exposeRange(expose('brightness', { attributes: { min: 0, max: 254 } }))).toEqual({ min: 0, max: 254 });
    expect(exposeRange(expose('brightness', { attributes: { value_min: 1, value_max: 100 } }))).toEqual({ min: 1, max: 100 });
    expect(exposeRange(expose('temperature'))).toBeNull();
  });
  test('numeric presets use their names and drop out-of-range sentinels', () => {
    const ct = expose('color_temp', { attributes: { min: 150, max: 500 }, values: { warm: 454, previous: 65535 } });
    expect(exposePresets(ct)).toEqual([{ label: 'Warm', value: 454 }]);
  });
  test('enum presets use their values', () => {
    const e = expose('power_on_behavior', { type: 'enum' as never, values: { '0': 'off', '1': 'previous_state' } });
    expect(exposePresets(e)).toEqual([{ label: 'Off', value: 'off' }, { label: 'Previous state', value: 'previous_state' }]);
  });
});

describe('controlOf', () => {
  test('picks the input for writable exposes only', () => {
    expect(controlOf(expose('state', { type: 'binary' as never, access_mode: 'readwrite' as never }))).toBe('switch');
    expect(controlOf(expose('brightness', { access_mode: 'readwrite' as never, attributes: { min: 0, max: 254 } }))).toBe('slider');
    expect(controlOf(expose('duration', { access_mode: 'write' as never }))).toBe('number');
    expect(controlOf(expose('volume', { type: 'enum' as never, access_mode: 'write' as never, values: { '0': 'low' } }))).toBe('choice');
    expect(controlOf(expose('temperature'))).toBe('none');
  });
});
