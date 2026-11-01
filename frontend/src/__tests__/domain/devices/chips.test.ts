import { describe, expect, test } from '@jest/globals';
import { BATTERY_WITHOUT_LEVEL_LABEL, LINK_QUALITY_TITLE, MAINS_POWER_LABEL, WIFI_SIGNAL_TITLE, linkChip, powerChip } from '@/domain/devices/chips';
import { device, expose } from '../../helpers/devices';

describe('device status chips', () => {
  test('link quality uses one formatter and the shared warning threshold', () => {
    expect(linkChip(device('a', {}, [expose('linkquality', { data: 62 })]))).toEqual({
      icon: 'signal',
      label: '62 LQI',
      tone: 'default',
      title: LINK_QUALITY_TITLE,
    });
    expect(linkChip(device('b', {}, [expose('linkquality', { data: 49 })]))?.tone).toBe('warn');
  });

  test('Wi-Fi RSSI is shown only when link quality is absent', () => {
    const withRssi = device('a', {}, [expose('rssi', { data: -60 })]);
    expect(linkChip(withRssi)).toEqual({ icon: 'wifi', label: '-60 dBm', tone: 'default', title: WIFI_SIGNAL_TITLE });
    const combined = device('b', {}, [expose('linkquality', { data: 80 }), expose('rssi', { data: -60 })]);
    expect(linkChip(combined)?.icon).toBe('signal');
  });

  test('battery percent and low-battery warning are formatted centrally', () => {
    expect(powerChip(device('a', { power_source: 'battery' }, [expose('battery', { data: 15 })]))).toEqual({
      icon: 'batteryLow',
      label: '15 %',
      tone: 'warn',
    });
    expect(powerChip(device('b', { power_source: 'battery' }, [expose('battery', { data: null })]))).toEqual({
      icon: 'battery',
      label: BATTERY_WITHOUT_LEVEL_LABEL,
      tone: 'default',
    });
  });

  test('mains and absent power sources', () => {
    expect(powerChip(device('a'))).toEqual({ label: MAINS_POWER_LABEL, tone: 'muted' });
    expect(powerChip(device('b', { power_source: '' }))).toBeNull();
  });
});
