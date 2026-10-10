/*
 * Sign-in session (spec 007 US-08, FR-11). Sign-in is an OAuth redirect: POST auth/login
 * returns the provider URL, the provider returns to /?auth=success with a session cookie, and
 * GET auth/me confirms the user. The previous UI's behaviour is kept; only the screens change.
 */
import { computed, ref } from 'vue';
import type { RouteLocationNormalizedLoaded } from 'vue-router';
import { store } from '@/store';
import { getOAuthUrl, logout, restoreSession } from '@/services/auth.service';
import { navigatePostLogin, navigatePostLogout } from '@/router/navigation';
import { User } from '@/types/auth.type';

export const AUTH_RETURN_PARAM = 'auth';

const signingIn = ref(false);
const error = ref('');

export function useSession() {
  const isAuthenticated = computed(() => store.getters['auth/isAuthenticated']() as boolean);
  const user = computed(() => store.getters['auth/user']() as User);

  /** Leaves the app for the OAuth provider. */
  async function signIn() {
    if (signingIn.value) return;
    signingIn.value = true;
    error.value = '';
    try {
      window.location.assign(await getOAuthUrl());
    } catch {
      error.value = 'Could not reach the hub to start sign-in. Check that it is running and try again.';
      signingIn.value = false;
    }
  }

  /** Completes the provider redirect (/?auth=success). Returns true when it handled one. */
  async function completeOAuthReturn(route?: RouteLocationNormalizedLoaded): Promise<boolean> {
    const params = new URLSearchParams(window.location.search);
    if (params.get(AUTH_RETURN_PARAM) !== 'success') return false;
    window.history.replaceState({}, document.title, window.location.pathname);
    signingIn.value = true;
    const session = await restoreSession();
    signingIn.value = false;
    if (session?.isAuthenticated) {
      store.dispatch('auth/loginUser', session);
      navigatePostLogin(route, true);
    } else {
      error.value = 'Sign-in did not complete. Please try again.';
    }
    return true;
  }

  /**
   * Resumes a still-valid hub session when this browser has no stored sign-in (cleared
   * storage, new tab after expiry of local state). Silent when there is none.
   */
  async function resumeSession(route?: RouteLocationNormalizedLoaded) {
    const session = await restoreSession();
    if (session?.isAuthenticated && !isAuthenticated.value) {
      store.dispatch('auth/loginUser', session);
      navigatePostLogin(route, true);
    }
  }

  async function signOut() {
    if (!(await logout())) {
      store.dispatch('alerts/showError', 'Sign-out failed. Check the connection to the hub and try again.');
      return;
    }
    store.dispatch('auth/logoutUser');
    navigatePostLogout();
  }

  return { isAuthenticated, user, signingIn, error, signIn, completeOAuthReturn, resumeSession, signOut };
}
