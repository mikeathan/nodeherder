import { RouteName } from '@/types/router';

/** Presentation-only route names; never persist detail URLs or household identifiers. */
const DEFAULT_VIEWS: readonly RouteName[] = [
  RouteName.Panel, RouteName.Overview, RouteName.GroupDashboard, RouteName.Devices,
  RouteName.DeviceList, RouteName.Viewer, RouteName.Assistant, RouteName.ConsoleViewer, RouteName.Settings,
];

export function isDefaultView(value: unknown): value is RouteName {
  return DEFAULT_VIEWS.includes(value as RouteName);
}

export function defaultView(value: unknown): RouteName {
  return isDefaultView(value) ? value : RouteName.GroupDashboard;
}

/** Explicit internal destinations take priority over the browser's startup preference. */
export function postLoginTarget(redirect: unknown, preferred: unknown): string | { name: RouteName } {
  if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//') && redirect !== '/') return redirect;
  return { name: defaultView(preferred) };
}
