/* Main navigation (single source for the sidebar, header titles and keyboard order). */
import { RouteName } from '@/types/router';
import { IconName } from '@/components/ui/icons';

export type NavItem = {
  route: RouteName;
  label: string;
  icon: IconName;
  /** Other routes that belong to this section (keeps the item highlighted). */
  includes?: RouteName[];
};

export const NAV_ITEMS: readonly NavItem[] = [
  { route: RouteName.Overview, label: 'Overview', icon: 'overview' },
  { route: RouteName.GroupDashboard, label: 'Home', icon: 'home' },
  { route: RouteName.Devices, label: 'Devices', icon: 'devices', includes: [RouteName.DeviceView, RouteName.DeviceList, RouteName.DevicePage] },
  { route: RouteName.Connections, label: 'Connections', icon: 'connections', includes: [RouteName.Connection, RouteName.ConnectionSetup, RouteName.ConnectionProvider, RouteName.ConnectionHelp,RouteName.ConnectionNotFound] },
  { route: RouteName.Viewer, label: 'Automations', icon: 'automation', includes: [RouteName.Editor, RouteName.Creator] },
  { route: RouteName.Assistant, label: 'Assistant', icon: 'assistant' },
  { route: RouteName.ConsoleViewer, label: 'Console', icon: 'console' },
  { route: RouteName.Settings, label: 'Settings', icon: 'settings' },
];

export function isNavActive(item: NavItem, current: unknown): boolean {
  return item.route === current || (item.includes ?? []).includes(current as RouteName);
}
