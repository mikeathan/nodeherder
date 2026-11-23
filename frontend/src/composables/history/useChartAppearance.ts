import { computed, onScopeDispose, ref, type ComputedRef } from 'vue';
import type { Expose } from '@/types/device';
import { exposeKind } from '@/domain/exposes';
import { useThemeSettings } from '@/composables/useThemeSettings';
import type { ChartAppearance } from '@/components/chart/history/options';

/** Tokens the charts read; one list so a new chart cannot forget one. */
const APPEARANCE_TOKENS = {
  muted: '--nh-text-muted',
  grid: '--nh-chart-grid',
  surface: '--nh-surface',
  text: '--nh-text',
  border: '--nh-border-strong',
} as const;

function readToken(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

/**
 * The theme values a chart needs, read once per theme change.
 *
 * Chart.js cannot consume CSS variables, so every chart has to resolve tokens and the
 * reduced-motion preference at render time. Doing that inline in each component duplicated the
 * token list and the "read to subscribe" trick (`void settings.value`) that makes the computed
 * recompute when the theme changes.
 *
 * @param expose Optional expose whose kind picks the series colour; omit for a neutral accent.
 * @param compact Whether the caller is rendering a compact chart. Pass a getter (`() =>
 * props.compact`) when the flag can change with the viewport: a one-time boolean would freeze
 * axis visibility and animation at the width the chart first mounted.
 * @returns A computed appearance object, recomputed when the theme settings or the OS motion
 * preference change.
 */
export function useChartAppearance(
  expose?: Expose,
  compact: boolean | (() => boolean) = false
): ComputedRef<ChartAppearance> {
  const { mode, settings } = useThemeSettings();
  // One query, observed for the component's lifetime: a mid-session OS change must invalidate the
  // chart, not only the first read.
  const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
  const reducedMotion = ref(motionQuery.matches);
  const onChange = (event: MediaQueryListEvent) => {
    reducedMotion.value = event.matches;
  };
  motionQuery.addEventListener('change', onChange);
  onScopeDispose(() => motionQuery.removeEventListener('change', onChange));

  return computed<ChartAppearance>(() => {
    // Referencing these is the subscription: the computed re-evaluates when the theme changes.
    void settings.value;
    return {
      color: readToken(
        expose ? `--nh-kind-${exposeKind(expose.name)}` : '--nh-accent'
      ),
      muted: readToken(APPEARANCE_TOKENS.muted),
      grid: readToken(APPEARANCE_TOKENS.grid),
      surface: readToken(APPEARANCE_TOKENS.surface),
      text: readToken(APPEARANCE_TOKENS.text),
      border: readToken(APPEARANCE_TOKENS.border),
      dark: mode.value === 'dark',
      compact: typeof compact === 'function' ? compact() : compact,
      reducedMotion: reducedMotion.value,
      fontFamily: getComputedStyle(document.body).fontFamily,
      series: {
        power: readToken('--nh-kind-power'),
        temp: readToken('--nh-kind-temp'),
        contact: readToken('--nh-kind-contact'),
      },
    };
  });
}
