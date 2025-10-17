import { computed } from 'vue';
import { store } from '@/store';

export function useAuth() {
  const isAuthenticated = computed(() => store.getters['auth/isAuthenticated']());
  const user = computed(() => store.getters['auth/user']());
  return { isAuthenticated, user };
}
