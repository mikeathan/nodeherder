import { describe, expect, test } from '@jest/globals';
import {
  DEVICE_ABOUT_MISSING,
  DeviceTabIds,
  adjacentDeviceTab,
  deviceAboutRows,
  deviceCategorySections,
  deviceTabLabel,
  deviceTabQuery,
  deviceTabs,
  lastSeenLabel,
  resolveDeviceTab,
} from '@/domain/devices/page';
import { device, expose } from '../../helpers/devices';

describe('device page model', () => {
  test('tabs resolve safely and keep the default out of the URL', () => {
    expect(resolveDeviceTab('history')).toBe(DeviceTabIds.History);
    expect(resolveDeviceTab('other')).toBe(DeviceTabIds.Controls);
    expect(deviceTabQuery(DeviceTabIds.Controls)).toBeUndefined();
    expect(deviceTabQuery(DeviceTabIds.About)).toBe(DeviceTabIds.About);
    expect(adjacentDeviceTab(deviceTabs(undefined), 0, -1).id).toBe(DeviceTabIds.About);
  });

  test('presence is offered only to devices that can map coordinates', () => {
    const plain = deviceTabs(device('plain'));
    expect(plain.map((tab) => tab.id)).not.toContain(DeviceTabIds.Presence);
    // A withheld tab in a shared link falls back to Controls instead of an empty panel.
    expect(resolveDeviceTab(DeviceTabIds.Presence, plain)).toBe(DeviceTabIds.Controls);

    const spatial = device('radar', { provider: 'esphome' }, [
      expose('target_x', { type: 'numeric' as never, unit: 'mm', data: 100 }),
      expose('target_y', { type: 'numeric' as never, unit: 'mm', data: 200 }),
      expose('target_active', { type: 'binary' as never, data: true }),
    ]);
    const capable = deviceTabs(spatial);
    expect(capable.map((tab) => tab.id)).toEqual([
      DeviceTabIds.Controls,
      DeviceTabIds.Presence,
      DeviceTabIds.History,
      DeviceTabIds.Settings,
      DeviceTabIds.About,
    ]);
    expect(resolveDeviceTab(DeviceTabIds.Presence, capable)).toBe(DeviceTabIds.Presence);
    // Keyboard navigation wraps within the tabs actually shown.
    expect(adjacentDeviceTab(capable, 4, 1).id).toBe(DeviceTabIds.Controls);
    expect(adjacentDeviceTab(capable, 1, 1).id).toBe(DeviceTabIds.History);
  });

  test('ESPHome uses hub-settings wording', () => {
    expect(deviceTabLabel(DeviceTabIds.Settings, true)).toBe('Hub settings');
    expect(deviceTabLabel(DeviceTabIds.Settings, false)).toBe('Settings');
  });

  test('About includes allowlisted provider fields only for ESPHome', () => {
    const provider = device('id', {
      provider: 'esphome',
      connection_type: 'esphome',
      description: 'Radar',
      properties: { manufacturer: 'Everything', firmware_version: '2026.10.0' },
    });
    const rows = deviceAboutRows(provider);
    expect(rows.find((row) => row.key === 'manufacturer')?.value).toBe('Everything');
    expect(rows.find((row) => row.key === 'project_name')?.value).toBe(DEVICE_ABOUT_MISSING);
    expect(rows.at(-1)).toMatchObject({ label: 'Last report', relative: true });
    expect(lastSeenLabel(provider)).toBe('Last report');
    expect(deviceAboutRows(device('legacy')).some((row) => row.key === 'manufacturer')).toBe(false);
  });

  test('category sections filter without changing entity identity', () => {
    const value = device('id', {}, [
      expose('temperature', { category: 'measurement' as never }),
      expose('state', { category: 'measurement' as never, type: 'binary' as never, access_mode: 'readwrite' as never }),
      expose('child_lock', { category: 'config' as never }),
      expose('linkquality', { category: 'diagnostic' as never }),
    ]);
    const sections = deviceCategorySections(value);
    expect(sections.measurements.map((item) => item.name)).toEqual(['temperature', 'state']);
    expect(sections.readings.map((item) => item.name)).toEqual(['temperature']);
    expect(sections.controls.map((item) => item.name)).toEqual(['state']);
    expect(sections.configuration.map((item) => item.name)).toEqual(['child_lock']);
    expect(sections.diagnostics.map((item) => item.name)).toEqual(['linkquality']);
  });
});
