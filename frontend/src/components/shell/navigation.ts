/* Main navigation (single source for the sidebar, header titles and keyboard order). */
import { RouteName } from '@/types/router';
import { IconName } from '@/components/ui/icons';

export type NavItem = {
  route: RouteName;
  label: string;
  icon: IconName;
  /** Other routes that belong to this section (keeps the item highlighted). */
  includes?: RouteName[];
  /**
   * When set, the item is only offered while the predicate holds. Used for capability-gated
   * sections so a hub with no source shows an empty section rather than a broken one.
   *
   * Wireless observation used to be a nav item and no longer is: it is an integration, reached from
   * its catalog card or a Home shortcut, so the route stays without a sidebar entry.
   */
  when?: () => boolean;
};

export const NAV_ITEMS: readonly NavItem[] = [
  { route: RouteName.Overview, label: 'Overview', icon: 'overview' },
  { route: RouteName.GroupDashboard, label: 'Home', icon: 'home' },
  { route: RouteName.Devices, label: 'Devices', icon: 'devices', includes: [RouteName.DeviceView, RouteName.DeviceList, RouteName.DevicePage] },
  { route: RouteName.Integrations, label: 'Integrations', icon: 'connections', includes: [RouteName.Integration, RouteName.IntegrationSetup, RouteName.IntegrationProvider, RouteName.IntegrationHelp,RouteName.IntegrationNotFound] },
  { route: RouteName.Viewer, label: 'Automations', icon: 'automation', includes: [RouteName.Editor, RouteName.Creator] },
  { route: RouteName.Assistant, label: 'Assistant', icon: 'assistant' },
  { route: RouteName.ConsoleViewer, label: 'Console', icon: 'console' },
  { route: RouteName.Settings, label: 'Settings', icon: 'settings' },
];

/** Items offered in the current hub state; capability-gated entries are filtered here. */
export function availableNavItems(items: readonly NavItem[] = NAV_ITEMS): readonly NavItem[] {
  return items.filter((item) => !item.when || item.when());
}

export function isNavActive(item: NavItem, current: unknown): boolean {
  return item.route === current || (item.includes ?? []).includes(current as RouteName);
}
