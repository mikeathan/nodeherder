import {
  SETTINGS_GROUP_LABELS,
  SETTINGS_SECTIONS,
  groupSettingsSections,
  resolveSettingsSection,
  sectionQueryValue,
} from '@/components/hub/settings/sections';

test('sections use stable literal ids in rail order', () => {
  expect(SETTINGS_SECTIONS.map((section) => section.id)).toEqual([
    'appearance',
    'device-defaults',
    'storage',
    'logger',
    'mcp',
    'assistant',
  ]);
});

test('every section names its own icon so the rail never reuses a nav or integration glyph', () => {
  const icons = SETTINGS_SECTIONS.map((section) => section.icon);
  expect(new Set(icons).size).toBe(icons.length);
  // These glyphs already mean Integrations, Console or Assistant elsewhere in the shell.
  expect(icons).not.toEqual(expect.arrayContaining(['schedule', 'chart', 'console', 'secure', 'assistant', 'connections']));
});

test('appearance is browser scoped and hub sections are grouped after it', () => {
  const groups = groupSettingsSections();
  expect(groups.map((group) => group.label)).toEqual([SETTINGS_GROUP_LABELS.browser, SETTINGS_GROUP_LABELS.hub]);
  expect(groups[0].sections.map((section) => section.id)).toEqual(['appearance']);
  expect(groups[1].sections.map((section) => section.id)).toEqual(['device-defaults', 'storage', 'logger', 'mcp', 'assistant']);
});

test('retired ids resolve to the merged storage section', () => {
  expect(resolveSettingsSection('history').id).toBe('storage');
  expect(resolveSettingsSection('system').id).toBe('storage');
  expect(resolveSettingsSection('mcp-server').id).toBe('mcp');
});

test('known ids resolve exactly and the first section keeps a clean URL', () => {
  expect(resolveSettingsSection('logger').id).toBe('logger');
  expect(sectionQueryValue(SETTINGS_SECTIONS[0])).toBeUndefined();
  expect(sectionQueryValue(SETTINGS_SECTIONS[2])).toBe('storage');
});

test('missing, empty, repeated and unknown queries fall back to the first section', () => {
  expect(resolveSettingsSection(undefined).id).toBe('appearance');
  expect(resolveSettingsSection('').id).toBe('appearance');
  expect(resolveSettingsSection(['logger', 'mcp']).id).toBe('logger');
  expect(resolveSettingsSection([]).id).toBe('appearance');
  expect(resolveSettingsSection('nope').id).toBe('appearance');
  expect(resolveSettingsSection(['logger']).id).toBe('logger');
});
