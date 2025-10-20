import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { store } from '@/store';
import { login, logout } from '@/services/auth.service';
import { navigatePostLogin, navigatePostLogout } from '@/router/navigation';
import { UserSession } from '@/types/auth.type';

export function useAuth() {
  const route = useRoute();

  const isAuthenticated = computed(() => store.getters['auth/isAuthenticated']());
  const user = computed(() => store.getters['auth/user']());

  const signIn = async () => {
    const userSession: UserSession = await login();
    console.log('User session after login:', userSession);
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

  return {
    isAuthenticated,
    user,
    signIn,
    signOut,
  };
}
