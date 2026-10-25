import { NAV_ITEMS, isNavActive } from '@/components/shell/navigation';
import { RouteName } from '@/types/router';
test('Connections follows Devices and owns only its feature routes', () => {
  const index = NAV_ITEMS.findIndex((i) => i.route === RouteName.Connections);
  expect(NAV_ITEMS[index - 1].route).toBe(RouteName.Devices);
  expect(isNavActive(NAV_ITEMS[index], RouteName.ConnectionHelp)).toBe(true);
  expect(isNavActive(NAV_ITEMS[index], RouteName.GroupDashboard)).toBe(false);
});
