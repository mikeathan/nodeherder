import { computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { store } from '@/store';
import { login, logout, restoreSession } from '@/services/auth.service';
import { navigatePostLogin, navigatePostLogout } from '@/router/navigation';
import { UserSession } from '@/types/auth.type';

export function useAuth() {
  const route = useRoute();

  const isAuthenticated = computed(() => store.getters['auth/isAuthenticated']());
  const user = computed(() => store.getters['auth/user']());

  const signIn = async () => {
    const userSession: UserSession = await login();
    if (!userSession.isAuthenticated) {
      console.error('Sign-in failed: User is not authenticated');
      // TODO:
      // Show toast notification
      return;
    }
    console.log('Sign-in successful:', userSession);

    store.dispatch('auth/loginUser', userSession);
    navigatePostLogin(route, userSession.isAuthenticated);
  };

  const signOut = async () => {
    const success = await logout();
    if (!success) {
      console.error('Logout failed');
      return;
    }

    store.dispatch('auth/logoutUser');
    navigatePostLogout();
  };

  onMounted(async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const authStatus = urlParams.get('auth');

    if (authStatus === 'success') {
      // Clear the URL parameter
      const newUrl = window.location.pathname;
      window.history.replaceState({}, document.title, newUrl);

      try {
        const session = await restoreSession();
        if (session && session.isAuthenticated) {
          store.dispatch('auth/loginUser', session);
          navigatePostLogin(route, true);
        } else {
          console.error('OAuth return: session not authenticated');
        }
      } catch (e) {
        console.error('OAuth return: failed to restore session', e);
      }
    }
  });

  return {
    isAuthenticated,
    user,
    signIn,
    signOut,
  };
}
