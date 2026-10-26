import { NAV_ITEMS, isNavActive } from '@/components/shell/navigation';
import { RouteName } from '@/types/router';
test('Integrations follows Devices and owns only its feature routes', () => {
  const index = NAV_ITEMS.findIndex((i) => i.route === RouteName.Integrations);
  expect(NAV_ITEMS[index - 1].route).toBe(RouteName.Devices);
  expect(isNavActive(NAV_ITEMS[index], RouteName.IntegrationHelp)).toBe(true);
  expect(isNavActive(NAV_ITEMS[index], RouteName.GroupDashboard)).toBe(false);
});
