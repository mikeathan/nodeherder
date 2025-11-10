import { router } from '@/router/index';
import { RouteName } from '@/types/router';
import type { RouteLocationNormalizedLoaded } from 'vue-router';

type RouteParams = {
  [RouteName.Editor]: { id: string };
  [RouteName.DevicePage]: { id: string };
  [RouteName.DeviceView]: { id: string };
  [key: string]: Record<string, any> | undefined;
};

export function goTo<N extends RouteName>(name: N, params?: RouteParams[N]) {
  router.push({ name, params });
}

export function goToPath(path: string) {
  router.push(path);
}

export function getPathForRoute(name: RouteName): string {
  const route = router.resolve({ name });
  return route.href;
}

function resolveAuthRoute(isAuthenticated: boolean, route?: RouteLocationNormalizedLoaded): string {
  if (isAuthenticated) {
    if (route) {
      const redirectPath = route.query.redirect as string;

      // If redirect exists and is not the login page itself, use it
      if (redirectPath && redirectPath !== '/' && redirectPath !== RouteName.Login) {
        return redirectPath;
      }
    }

    // Default authenticated route
    return getPathForRoute(RouteName.GroupDashboard);
  }

  // User is not authenticated - go to login
  return getPathForRoute(RouteName.Login);
}

export function navigatePostLogin(route?: RouteLocationNormalizedLoaded, isAuthenticated = true) {
  const path = resolveAuthRoute(isAuthenticated, route);
  goToPath(path);
}

export function navigatePostLogout() {
  const path = resolveAuthRoute(false);
  goToPath(path);
}
