import { RouteLocationNormalizedLoaded } from 'vue-router';
import { store } from '@/store';

/**
 * Default routes for authentication flows
 */
const ROUTES = {
  Groupdashboard: '/groupdashboard',
  Login: '/login',
} as const;

/**
 * Resolves the redirect path based on authentication state
 * @param route - The current route object (optional, only needed for login flow)
 * @returns The path to redirect to
 */
export function resolveAuthRoute(route?: RouteLocationNormalizedLoaded): string {
  const isAuthenticated = store.getters['auth/isAuthenticated']();

  if (isAuthenticated) {
    // User is logged in - check if there's a redirect from login flow
    if (route) {
      const redirectPath = route.query.redirect as string;

      // If redirect exists and is not the login page itself, use it
      if (redirectPath && redirectPath !== '/' && redirectPath !== ROUTES.Login) {
        return redirectPath;
      }
    }

    // Default authenticated route
    return ROUTES.Groupdashboard;
  }

  // User is not authenticated - go to login
  return ROUTES.Login;
}

/**
 * Resolves the redirect path after successful login
 * @param route - The current route object
 * @returns The path to redirect to after login
 */
export function resolvePostLoginRoute(route: RouteLocationNormalizedLoaded): string {
  return resolveAuthRoute(route);
}

/**
 * Gets the path to redirect to after logout
 * @returns The path to redirect to after logout
 */
export function resolvePostLogoutRoute(): string {
  return resolveAuthRoute();
}
