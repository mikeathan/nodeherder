/*
 * Appearance preferences (spec 007 FR-02). One shared state for the whole app: the
 * customiser panel and Settings → Appearance edit the same settings, which are validated
 * (theme/theme-settings.ts), stored in localStorage and applied to <html> as attributes and
 * token overrides.
 */
import { computed, readonly, ref, watch } from 'vue';
import {
  DEFAULT_THEME_SETTINGS,
  documentTheme,
  LEGACY_THEME_KEY,
  loadThemeSettings,
  parseThemeSettings,
  resolveMode,
  THEME_STORAGE_KEY,
  ThemeSettings,
} from '@/theme/theme-settings';
import { safeStorage } from '@/utils/storage';

const DARK_QUERY = '(prefers-color-scheme: dark)';

const settings = ref<ThemeSettings>(loadThemeSettings(safeStorage));
const prefersDark = ref(false);
let started = false;

function applyToDocument() {
  if (typeof document === 'undefined') return;
  const root = document.documentElement;
  const { attributes, properties } = documentTheme(settings.value, prefersDark.value);
  for (const [name, value] of Object.entries(attributes)) root.setAttribute(name, value);
  for (const [name, value] of Object.entries(properties)) {
    if (value === null) root.style.removeProperty(name);
    else root.style.setProperty(name, value);
  }
}

/** Applies stored settings and starts following them; safe to call more than once. */
export function startThemeSettings() {
  if (started) return;
  started = true;
  const media = typeof window !== 'undefined' && window.matchMedia ? window.matchMedia(DARK_QUERY) : null;
  prefersDark.value = media?.matches ?? false;
  media?.addEventListener?.('change', (e) => (prefersDark.value = e.matches));
  applyToDocument();
  watch([settings, prefersDark], applyToDocument, { deep: true });
  watch(
    settings,
    (value) => {
      safeStorage.setItem(THEME_STORAGE_KEY, JSON.stringify(value));
      safeStorage.removeItem(LEGACY_THEME_KEY);
    },
    { deep: true }
  );
}

export function useThemeSettings() {
  startThemeSettings();
  const mode = computed(() => resolveMode(settings.value.mode, prefersDark.value));
  return {
    settings: readonly(settings),
    /** Light or dark as currently shown (resolves "system"). */
    mode,
    update(patch: Partial<ThemeSettings>) {
      settings.value = parseThemeSettings({ ...settings.value, ...patch });
    },
    toggleMode() {
      settings.value = parseThemeSettings({ ...settings.value, mode: mode.value === 'dark' ? 'light' : 'dark' });
    },
    reset() {
      settings.value = { ...DEFAULT_THEME_SETTINGS };
    },
  };
}
