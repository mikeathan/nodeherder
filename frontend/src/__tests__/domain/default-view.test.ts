import { defaultView, isDefaultView, postLoginTarget } from '@/domain/default-view';
import { RouteName } from '@/types/router';

describe('browser default view policy', () => {
  it.each([RouteName.Panel, RouteName.Overview, RouteName.GroupDashboard, RouteName.Devices, RouteName.DeviceList, RouteName.Viewer, RouteName.Assistant, RouteName.ConsoleViewer, RouteName.Settings])('accepts main view %s', (name) => {
    expect(isDefaultView(name)).toBe(true);
    expect(defaultView(name)).toBe(name);
  });
  it.each([null, undefined, '', 'removed-route', '/panel', RouteName.Login, RouteName.Creator, RouteName.Editor, RouteName.DevicePage, RouteName.DeviceView, {}, ['panel']])('falls back for unsupported value %p', (value) => {
    expect(isDefaultView(value)).toBe(false);
    expect(defaultView(value)).toBe(RouteName.GroupDashboard);
  });
  it('uses a pinned view only when there is no explicit destination', () => {
    expect(postLoginTarget(undefined, RouteName.Panel)).toEqual({ name: RouteName.Panel });
    expect(postLoginTarget('/settings', RouteName.Panel)).toBe('/settings');
    expect(postLoginTarget('/devicepage/123?tab=metrics', RouteName.Panel)).toBe('/devicepage/123?tab=metrics');
  });
  it.each(['/', 'login', '', undefined, ['/panel'], '//example.com', 'https://example.com'])('ignores unsuitable login redirect %p', (redirect) => {
    expect(postLoginTarget(redirect, null)).toEqual({ name: RouteName.GroupDashboard });
  });
});
