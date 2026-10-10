import { describe, expect, test } from '@jest/globals';
import { batteryLevel, deviceStatus, hasLowBattery, hasWeakLink, isBatteryPowered, linkQuality, resolveProtocol } from '@/domain/devices';
import { device, expose } from '../helpers/devices';

describe('protocol registry', () => {
  test('known protocols and their supported bridge operations', () => {
    expect(resolveProtocol('mqtt')).toMatchObject({ label: 'Zigbee', supports: { rename: true, interview: true, remove: true }, generic: false });
    expect(resolveProtocol('http')).toMatchObject({ label: 'Wi-Fi · HTTP', supports: { rename: false, interview: false, remove: false } });
  });
  test('unknown protocols get a generic entry with no bridge operations (AC-08)', () => {
    expect(resolveProtocol('matter')).toEqual({ id: 'matter', label: 'matter', supports: { rename: false, interview: false, remove: false }, generic: true });
    expect(resolveProtocol(undefined).label).toBe('Unknown');
  });
});

describe('device health', () => {
  test('power source is free text', () => {
    expect(isBatteryPowered(device('a', { power_source: 'battery' }))).toBe(true);
    expect(isBatteryPowered(device('b', { power_source: 'mains (single phase)' }))).toBe(false);
    expect(isBatteryPowered(device('c', { power_source: '' }))).toBe(false);
  });
  test('battery from battery or battpercentage; link quality; thresholds', () => {
    expect(batteryLevel(device('a', {}, [expose('battpercentage', { data: 15 })]))).toBe(15);
    expect(batteryLevel(device('a', {}, [expose('battery', { data: null })]))).toBeNull();
    expect(hasLowBattery(device('a', {}, [expose('battery', { data: 19 })]))).toBe(true);
    expect(hasLowBattery(device('a', {}, [expose('battery', { data: 20 })]))).toBe(false);
    expect(linkQuality(device('a', {}, [expose('linkquality', { data: 49 })]))).toBe(49);
    expect(hasWeakLink(device('a', {}, [expose('linkquality', { data: 49 })]))).toBe(true);
    expect(hasWeakLink(device('a'))).toBe(false);
  });
  test('status: disabled wins, then availability', () => {
    expect(deviceStatus(device('a', { availability: 'offline' as never }), true)).toBe('disabled');
    expect(deviceStatus(device('a', { availability: 'offline' as never }), false)).toBe('offline');
    expect(deviceStatus(device('a', { availability: 'unknown' as never }), false)).toBe('unknown');
  });
});
