/*
 * Sign-in session (spec 008). POST auth/login selects local or Google access;
 * the redirect returns to /?auth=success with a cookie and auth/me confirms the user.
 */
import { computed, ref } from 'vue';
import type { RouteLocationNormalizedLoaded } from 'vue-router';
import { store } from '@/store';
import { getOAuthUrl, logout } from '@/services/auth.service';
import { navigatePostLogin, navigatePostLogout } from '@/router/navigation';
import { User } from '@/types/auth.type';

export const AUTH_RETURN_PARAM = 'auth';

const signingIn = ref(false);
const error = ref('');

export function useSession() {
  const isAuthenticated = computed(() => store.getters['auth/isAuthenticated']() as boolean);
  const user = computed(() => store.getters['auth/user']() as User);
  const validation = computed(() => store.getters['auth/validation']);

  /** Follows the hub's selected local/Google sign-in path. */
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
    await store.dispatch('auth/restoreSession');
    signingIn.value = false;
    if (isAuthenticated.value) {
      navigatePostLogin(route, true);
    } else {
      error.value = validation.value === 'unavailable'
        ? 'Could not reach the hub to check your session. Please retry.'
        : 'Sign-in did not complete. Please try again.';
    }
    return true;
  }

  /**
   * Resumes a still-valid hub session when this browser has no stored sign-in (cleared
   * storage, new tab after expiry of local state). Silent when there is none.
   */
  async function resumeSession(route?: RouteLocationNormalizedLoaded) {
    await store.dispatch('auth/restoreSession');
    if (isAuthenticated.value) {
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

  return { isAuthenticated, user, validation, signingIn, error, signIn, completeOAuthReturn, resumeSession, signOut };
}
