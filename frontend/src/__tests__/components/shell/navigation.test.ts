import { NAV_ITEMS, availableNavItems, isNavActive } from '@/components/shell/navigation';
import { RouteName } from '@/types/router';

test('Wireless is not a nav item', () => {
  // It is an integration, not a section of the application. A top-level entry implied the hub has a
  // wireless feature rather than one reader, putting it beside Device control and Automations as if
  // it were the same kind of thing.
  const routes = NAV_ITEMS.map((item) => item.route);
  expect(routes).not.toContain(RouteName.Wireless);

  // The route survives, so integration cards, Home shortcuts and deep links still reach the view.
  expect(RouteName.Wireless).toBeTruthy();
});

test('Integrations follows Devices', () => {
  const devices = NAV_ITEMS.findIndex((item) => item.route === RouteName.Devices);
  const integrations = NAV_ITEMS.findIndex((item) => item.route === RouteName.Integrations);
  expect(integrations).toBe(devices + 1);
});

test('Integrations owns only its feature routes', () => {
  const item = NAV_ITEMS.find((candidate) => candidate.route === RouteName.Integrations)!;
  expect(isNavActive(item, RouteName.IntegrationHelp)).toBe(true);
  expect(isNavActive(item, RouteName.GroupDashboard)).toBe(false);
});

test('capability-gated items are hidden when their predicate is false', () => {
  const items = [
    { route: RouteName.Devices, label: 'Devices', icon: 'devices' as const },
    { route: RouteName.Wireless, label: 'Wireless', icon: 'signal' as const, when: () => false },
  ];
  expect(availableNavItems(items).map((item) => item.route)).toEqual([RouteName.Devices]);
});

test('capability-gated items are shown when their predicate is true', () => {
  const items = [
    { route: RouteName.Wireless, label: 'Wireless', icon: 'signal' as const, when: () => true },
  ];
  expect(availableNavItems(items)).toHaveLength(1);
});

test('ungated items are always available', () => {
  expect(availableNavItems().length).toBe(NAV_ITEMS.length);
});
