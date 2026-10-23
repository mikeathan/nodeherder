import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import {
  ACCENTS,
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

  it('offers ten distinct accent swatches without the duplicate blue, brown or mustard', () => {
    expect(ACCENTS).toHaveLength(10);
    expect(new Set(ACCENTS).size).toBe(10);
    ['#0f8bd6', '#8a5a2b', '#c08a1e'].forEach((hex) => expect(ACCENTS).not.toContain(hex));
    expect(ACCENTS).toEqual(expect.arrayContaining(['#c92a2a', '#2b4fd8', '#495057']));
    ACCENTS.forEach((hex) => expect(parseThemeSettings({ accent: hex }).accent).toBe(hex));
  });

  it('keeps the surviving retro theme, adds the review picks and drops the rejected palettes', () => {
    expect(THEMES.find((t) => t.id === 'amber')?.name).toBe('Terminal Amber');
    expect(THEMES.filter((t) => ['crimson', 'emerald', 'blueprint'].includes(t.id)).map((t) => t.name))
      .toEqual(['Crimson', 'Emerald', 'Blueprint']);
    ['teak', 'avocado', 'plum', 'sunset', 'diner'].forEach((id) =>
      expect(THEMES.some((t) => t.id === id)).toBe(false));
    THEMES.forEach((t) => expect(t.swatch).toHaveLength(2));
    expect(new Set(THEMES.map((t) => t.swatch.join()))).toHaveProperty('size', THEMES.length);
  });

  it('defines every token for every theme in both modes', () => {
    const css = readFileSync(resolve(__dirname, '../../assets/styles/themes.css'), 'utf8');
    const tokens = ['--nh-bg', '--nh-surface', '--nh-surface-2', '--nh-border', '--nh-border-strong',
      '--nh-text', '--nh-text-muted', '--nh-accent', '--nh-accent-contrast', '--nh-nav-bg',
      '--nh-header-bg', '--dk-lamp', '--dk-climate', '--dk-glass'];
    for (const theme of THEMES) {
      for (const mode of ['light', 'dark']) {
        const block = css.match(new RegExp(`\\[data-theme='${theme.id}'\\]\\[data-mode='${mode}'\\] \\{([^}]*)\\}`))?.[1];
        expect(block).toBeDefined();
        tokens.forEach((token) => expect(block).toMatch(new RegExp(`${token}:\\s*[^;]+;`)));
      }
    }
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
