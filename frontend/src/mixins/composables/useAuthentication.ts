import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { store } from '@/store';
import { resolvePostLoginRoute, resolvePostLogoutRoute } from '@/utils/auth';

export function useAuth() {
  const router = useRouter();
  const route = useRoute();

  const isAuthenticated = computed(() => store.getters['auth/isAuthenticated']());
  const user = computed(() => store.getters['auth/user']());

  const login = (loginData: { user: any; token: string }) => {
    store.dispatch('auth/loginUser', loginData);
    const redirectPath = resolvePostLoginRoute(route);
    router.push(redirectPath);
  };

  // TODO:
  //Add a backend logout endpoint to invalidate/blacklist the JWT token for better security

  const logout = () => {
    store.dispatch('auth/logoutUser');
    const redirectPath = resolvePostLogoutRoute();
    router.push(redirectPath);
  };

  return {
    isAuthenticated,
    user,
    login,
    logout,
  };
}
