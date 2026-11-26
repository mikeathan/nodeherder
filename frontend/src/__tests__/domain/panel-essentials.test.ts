import { describe, expect, test } from '@jest/globals';
import { PANEL_ESSENTIAL_LIMIT, parsePanelEssentials, togglePanelEssential } from '@/domain/panel-essentials';

const item = (id: string) => ({ deviceId: `device-${id}`, expose: 'state', label: `Light ${id}` });

describe('panel night essentials', () => {
  test('parses only valid, unique entries and caps the list', () => {
    const parsed = parsePanelEssentials([
      { deviceId: 'a', expose: 'state', label: 'A' },
      { deviceId: 'a', expose: 'state', label: 'duplicate' },
      { deviceId: 'b', expose: 'brightness', label: 'B' },
      { deviceId: 1, expose: 'state', label: 'bad' },
      { deviceId: 'c', expose: 'state', label: 'C' },
      { deviceId: 'd', expose: 'state', label: 'D' },
      { deviceId: 'e', expose: 'state', label: 'E' },
    ]);
    expect(parsed.map((entry) => entry.label)).toEqual(['A', 'B', 'C', 'D']);
    expect(parsed).toHaveLength(PANEL_ESSENTIAL_LIMIT);
  });

  test('falls back safely for malformed data', () => {
    expect(parsePanelEssentials(null)).toEqual([]);
    expect(parsePanelEssentials(['nope'])).toEqual([]);
  });

  test('toggles selection without exceeding the limit', () => {
    let selected = togglePanelEssential([], item('a'));
    selected = togglePanelEssential(selected, item('b'));
    selected = togglePanelEssential(selected, item('a'));
    expect(selected.map((entry) => entry.label)).toEqual(['Light b']);

    let full = [item('1'), item('2'), item('3'), item('4')];
    full = togglePanelEssential(full, item('5'));
    expect(full).toHaveLength(PANEL_ESSENTIAL_LIMIT);
  });
});
