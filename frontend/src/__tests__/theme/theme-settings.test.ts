import {
  contrastOn,
  DEFAULT_THEME_SETTINGS,
  documentTheme,
  LEGACY_THEME_KEY,
  loadThemeSettings,
  parseThemeSettings,
  resolveMode,
  THEME_STORAGE_KEY,
  THEMES,
} from '@/theme/theme-settings';

const storage = (items: Record<string, string>) => ({ getItem: (k: string) => items[k] ?? null });

describe('parseThemeSettings', () => {
  it('returns defaults for junk', () => {
    expect(parseThemeSettings(null)).toEqual(DEFAULT_THEME_SETTINGS);
    expect(parseThemeSettings('dark')).toEqual(DEFAULT_THEME_SETTINGS);
  });

  it('keeps valid fields and replaces invalid ones individually', () => {
    expect(parseThemeSettings({ mode: 'dark', theme: 'nope', accent: '#ABCDEF', density: 'huge', radius: 99, fontScale: 80.4, effects: false })).toEqual({
      ...DEFAULT_THEME_SETTINGS,
      mode: 'dark',
      accent: '#abcdef',
      radius: 20,
      fontScale: 90,
      effects: false,
    });
    expect(parseThemeSettings({ accent: 'red', radius: NaN }).accent).toBe('theme');
  });

  it('knows every theme id exactly once', () => {
    const ids = THEMES.map((t) => t.id);
    expect(new Set(ids).size).toBe(ids.length);
    ids.forEach((id) => expect(parseThemeSettings({ theme: id }).theme).toBe(id));
  });
});

describe('loadThemeSettings', () => {
  it('prefers stored settings', () => {
    expect(loadThemeSettings(storage({ [THEME_STORAGE_KEY]: JSON.stringify({ theme: 'ember' }), [LEGACY_THEME_KEY]: 'dark' })).theme).toBe('ember');
  });

  it('migrates the legacy dark/light key', () => {
    expect(loadThemeSettings(storage({ [LEGACY_THEME_KEY]: 'dark' }))).toEqual({ ...DEFAULT_THEME_SETTINGS, mode: 'dark' });
  });

  it('survives missing or broken storage', () => {
    expect(loadThemeSettings(null)).toEqual(DEFAULT_THEME_SETTINGS);
    expect(loadThemeSettings(storage({ [THEME_STORAGE_KEY]: '{oops' }))).toEqual(DEFAULT_THEME_SETTINGS);
    expect(loadThemeSettings({ getItem: () => { throw new Error('denied'); } })).toEqual(DEFAULT_THEME_SETTINGS);
  });

  it('returns a copy that callers may mutate', () => {
    const s = loadThemeSettings(null);
    s.radius = 3;
    expect(DEFAULT_THEME_SETTINGS.radius).toBe(14);
  });
});

describe('documentTheme', () => {
  it('resolves system mode', () => {
    expect(resolveMode('system', true)).toBe('dark');
    expect(resolveMode('system', false)).toBe('light');
    expect(resolveMode('light', true)).toBe('light');
  });

  it('maps settings to attributes and token overrides', () => {
    expect(documentTheme(DEFAULT_THEME_SETTINGS, false)).toEqual({
      attributes: { 'data-mode': 'light', 'data-theme': 'fleece', 'data-density': 'comfortable', 'data-effects': 'on' },
      properties: { '--nh-r': '14', '--nh-scale': '1', '--nh-accent': null, '--nh-accent-contrast': null },
    });
    const custom = documentTheme({ ...DEFAULT_THEME_SETTINGS, accent: '#ffd43b', effects: false, fontScale: 110 }, true);
    expect(custom.attributes['data-effects']).toBe('off');
    expect(custom.properties).toMatchObject({ '--nh-scale': '1.1', '--nh-accent': '#ffd43b', '--nh-accent-contrast': '#111418' });
  });

  it('picks readable text on accents', () => {
    expect(contrastOn('#1b7fb0')).toBe('#ffffff');
    expect(contrastOn('#ffffff')).toBe('#111418');
    expect(contrastOn('#000000')).toBe('#ffffff');
  });
});
