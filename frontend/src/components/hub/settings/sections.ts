/**
 * Settings section registry: the single source of truth for what `/settings` can show,
 * which group it belongs to, and how a `?section=` value resolves.
 *
 * Ids are literals (never derived from titles) so retitling a section cannot silently
 * break a saved link. Sections load lazily to keep the initial settings bundle small.
 */
import { defineAsyncComponent, type Component } from 'vue';
import type { IconName } from '@/components/ui/icons';

/** Scope a section is saved in; drives the rail captions. */
export const SETTINGS_GROUPS = {
  browser: 'browser',
  hub: 'hub',
} as const;

export type SettingsGroup = (typeof SETTINGS_GROUPS)[keyof typeof SETTINGS_GROUPS];

export const SETTINGS_GROUP_LABELS: Readonly<Record<SettingsGroup, string>> = {
  browser: 'This browser',
  hub: 'Hub',
};

export type SettingsSectionId = 'appearance' | 'device-defaults' | 'storage' | 'logger' | 'mcp' | 'assistant';

export interface SettingsSection {
  id: SettingsSectionId;
  title: string;
  group: SettingsGroup;
  /** Rail glyph, chosen per section rather than reused from navigation or integrations. */
  icon: IconName;
  content: Component;
}

export const SETTINGS_SECTIONS: readonly SettingsSection[] = [
  {
    id: 'appearance',
    title: 'Appearance',
    group: SETTINGS_GROUPS.browser,
    icon: 'palette',
    content: defineAsyncComponent(() => import('@/components/hub/settings/AppearanceSection.vue')),
  },
  {
    id: 'device-defaults',
    title: 'Device defaults',
    group: SETTINGS_GROUPS.hub,
    icon: 'settingsDevices',
    content: defineAsyncComponent(() => import('@/components/hub/settings/DeviceDefaultsSection.vue')),
  },
  {
    id: 'storage',
    title: 'Storage & retention',
    group: SETTINGS_GROUPS.hub,
    icon: 'settingsStorage',
    content: defineAsyncComponent(() => import('@/components/hub/settings/StorageSettings.vue')),
  },
  {
    id: 'logger',
    title: 'Logger',
    group: SETTINGS_GROUPS.hub,
    icon: 'settingsLogger',
    content: defineAsyncComponent(() => import('@/components/hub/settings/LoggerSettings.vue')),
  },
  {
    id: 'mcp',
    title: 'MCP server',
    group: SETTINGS_GROUPS.hub,
    icon: 'settingsServer',
    content: defineAsyncComponent(() => import('@/components/hub/settings/MCPSettings.vue')),
  },
  {
    id: 'assistant',
    title: 'Assistant',
    group: SETTINGS_GROUPS.hub,
    icon: 'settingsAssistant',
    content: defineAsyncComponent(() => import('@/components/hub/settings/AssistantSettings.vue')),
  },
] as const;

/**
 * Retired ids that still appear in saved links and documentation. `history` and `system`
 * were merged into one section; both now open it.
 */
export const SETTINGS_SECTION_ALIASES: Readonly<Record<string, SettingsSectionId>> = {
  history: 'storage',
  system: 'storage',
  'mcp-server': 'mcp',
};

/** Sections in rail order, grouped for rendering. */
export function groupSettingsSections(
  sections: readonly SettingsSection[] = SETTINGS_SECTIONS
): { group: SettingsGroup; label: string; sections: SettingsSection[] }[] {
  const order = [SETTINGS_GROUPS.browser, SETTINGS_GROUPS.hub];
  return order.map((group) => ({
    group,
    label: SETTINGS_GROUP_LABELS[group],
    sections: sections.filter((section) => section.group === group),
  }));
}

/** Read a raw `?section=` value (which may be an array) as a single string. */
function readSectionQuery(query: unknown): string {
  if (typeof query === 'string') return query;
  if (Array.isArray(query) && typeof query[0] === 'string') return query[0];
  return '';
}

/**
 * Resolve a `?section=` query value to a section.
 *
 * Unknown and missing values fall back to the first section so a stale or misspelled link
 * still renders the page. Retired ids resolve through `SETTINGS_SECTION_ALIASES`.
 */
export function resolveSettingsSection(
  query: unknown,
  sections: readonly SettingsSection[] = SETTINGS_SECTIONS
): SettingsSection {
  const requested = readSectionQuery(query);
  const alias = SETTINGS_SECTION_ALIASES[requested];
  const id = alias ?? requested;
  return sections.find((section) => section.id === id) ?? sections[0];
}

/** The query value to store for a section; the first section keeps a clean URL. */
export function sectionQueryValue(section: SettingsSection): string | undefined {
  return section.id === SETTINGS_SECTIONS[0].id ? undefined : section.id;
}
