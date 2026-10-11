/*
 * Appearance preferences (spec 007 FR-02, AC-02, AC-18). Presentation only: never household
 * data (FE-05). Colour values live in src/assets/styles/themes.css (single source); this
 * module only knows theme ids, names and swatches for the picker.
 */
export type ThemeMode = 'system' | 'light' | 'dark';
export type Density = 'compact' | 'comfortable' | 'spacious';

export type ThemeDescriptor = { id: string; name: string; swatch: [string, string]; origin: 'hearth' | 'panel' | 'new' };

export const THEMES: readonly ThemeDescriptor[] = [
  { id: 'fleece', name: 'Fleece', swatch: ['#eef6fa', '#1b7fb0'], origin: 'new' },
  { id: 'daylight', name: 'Daylight', swatch: ['#f4f2ee', '#0f8bd6'], origin: 'hearth' },
  { id: 'teak', name: 'Teak', swatch: ['#f3ece1', '#b8611f'], origin: 'hearth' },
  { id: 'avocado', name: 'Avocado', swatch: ['#eef0e2', '#5e7d1e'], origin: 'hearth' },
  { id: 'lagoon', name: 'Lagoon', swatch: ['#13252c', '#6fd3e8'], origin: 'panel' },
  { id: 'ember', name: 'Ember', swatch: ['#24170f', '#ff8a3d'], origin: 'panel' },
  { id: 'frost', name: 'Frost', swatch: ['#182433', '#9cc9ff'], origin: 'panel' },
  { id: 'fjord', name: 'Fjord', swatch: ['#0f1a20', '#4fc3cf'], origin: 'new' },
  { id: 'plum', name: 'Plum', swatch: ['#f6f1f5', '#8e3b7f'], origin: 'new' },
  { id: 'graphite', name: 'Graphite', swatch: ['#1c1c1e', '#e8e8e8'], origin: 'new' },
];

export const ACCENTS: readonly string[] = ['#1b7fb0', '#0f8bd6', '#d9480f', '#2b8a3e', '#8e3b7f', '#0e7c86', '#c2255c', '#7048e8'];

export type ThemeSettings = {
  v: 1;
  mode: ThemeMode;
  theme: string;
  /** 'theme' = the theme's own accent, otherwise a #rrggbb colour. */
  accent: string;
  density: Density;
  radius: number;
  fontScale: number;
  effects: boolean;
};

export const DEFAULT_THEME_SETTINGS: ThemeSettings = Object.freeze({
  v: 1,
  mode: 'system',
  theme: 'fleece',
  accent: 'theme',
  density: 'comfortable',
  radius: 14,
  fontScale: 100,
  effects: true,
}) as ThemeSettings;

export const THEME_STORAGE_KEY = 'nodeherder_theme';
/** Written by the previous UI (services/theme.service.ts): 'dark' | 'light'. */
export const LEGACY_THEME_KEY = 'theme';

const HEX = /^#[0-9a-f]{6}$/i;
const oneOf = <T extends string>(v: unknown, allowed: readonly T[], fallback: T): T => (allowed.includes(v as T) ? (v as T) : fallback);
const clampInt = (v: unknown, min: number, max: number, fallback: number) =>
  typeof v === 'number' && Number.isFinite(v) ? Math.round(Math.min(max, Math.max(min, v))) : fallback;

/** Validates untrusted stored JSON field by field; anything invalid falls back to the default. */
export function parseThemeSettings(raw: unknown): ThemeSettings {
  const o = raw && typeof raw === 'object' ? (raw as Record<string, unknown>) : {};
  const d = DEFAULT_THEME_SETTINGS;
  return {
    v: 1,
    mode: oneOf(o.mode, ['system', 'light', 'dark'] as const, d.mode),
    theme: THEMES.some((t) => t.id === o.theme) ? (o.theme as string) : d.theme,
    accent: o.accent === 'theme' || (typeof o.accent === 'string' && HEX.test(o.accent)) ? (o.accent as string).toLowerCase() : d.accent,
    density: oneOf(o.density, ['compact', 'comfortable', 'spacious'] as const, d.density),
    radius: clampInt(o.radius, 0, 20, d.radius),
    fontScale: clampInt(o.fontScale, 90, 125, d.fontScale),
    effects: typeof o.effects === 'boolean' ? o.effects : d.effects,
  };
}

/** Reads stored settings, migrating the legacy dark/light key once. */
export function loadThemeSettings(storage: Pick<Storage, 'getItem'> | null): ThemeSettings {
  if (!storage) return { ...DEFAULT_THEME_SETTINGS };
  try {
    const stored = storage.getItem(THEME_STORAGE_KEY);
    if (stored) return parseThemeSettings(JSON.parse(stored));
    const legacy = storage.getItem(LEGACY_THEME_KEY);
    if (legacy === 'dark' || legacy === 'light') return { ...DEFAULT_THEME_SETTINGS, mode: legacy };
  } catch {
    /* unreadable storage or JSON: use defaults */
  }
  return { ...DEFAULT_THEME_SETTINGS };
}

export function resolveMode(mode: ThemeMode, prefersDark: boolean): 'light' | 'dark' {
  return mode === 'system' ? (prefersDark ? 'dark' : 'light') : mode;
}

/** Text colour that stays readable on the accent (WCAG relative luminance). */
export function contrastOn(hex: string): '#ffffff' | '#111418' {
  const n = parseInt(hex.slice(1), 16);
  const channel = (c: number) => {
    const s = c / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  const l = 0.2126 * channel((n >> 16) & 255) + 0.7152 * channel((n >> 8) & 255) + 0.0722 * channel(n & 255);
  return l > 0.36 ? '#111418' : '#ffffff';
}

export type DocumentTheme = { attributes: Record<string, string>; properties: Record<string, string | null> };

/** What to set on <html>: attributes select CSS rules, properties override tokens. */
export function documentTheme(s: ThemeSettings, prefersDark: boolean): DocumentTheme {
  const custom = s.accent !== 'theme';
  return {
    attributes: { 'data-mode': resolveMode(s.mode, prefersDark), 'data-theme': s.theme, 'data-density': s.density, 'data-effects': s.effects ? 'on' : 'off' },
    properties: {
      '--nh-r': String(s.radius),
      '--nh-scale': String(s.fontScale / 100),
      '--nh-accent': custom ? s.accent : null,
      '--nh-accent-contrast': custom ? contrastOn(s.accent) : null,
    },
  };
}
