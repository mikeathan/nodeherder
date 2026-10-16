import { router } from '@/router/index';
import { useDefaultView } from '@/composables/useDefaultView';
import { postLoginTarget } from '@/domain/default-view';
import { RouteName } from '@/types/router';
import type { RouteLocationNormalizedLoaded } from 'vue-router';

type RouteParams = {
  [RouteName.Editor]: { id: string };
  [RouteName.DevicePage]: { id: string };
  [RouteName.DeviceView]: { id: string };
  [key: string]: Record<string, string> | undefined;
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

export function navigatePostLogin(route?: RouteLocationNormalizedLoaded, isAuthenticated = true) {
  const target = isAuthenticated
    ? postLoginTarget(route?.query.redirect, useDefaultView().preferred.value)
    : { name: RouteName.Login };
  router.push(target);
}

export function navigatePostLogout() {
  goTo(RouteName.Login);
}
