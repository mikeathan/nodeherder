import { store } from '@/store';

export function useAlerts() {
  function showError(error?: string) {
    store.dispatch('alerts/showError', error);
  }

  function showSuccess(message?: string) {
    store.dispatch('alerts/showSuccess', message);
  }

  function showInfo(message?: string) {
    store.dispatch('alerts/showInfo', message);
  }

  function showWarning(message?: string) {
    store.dispatch('alerts/showWarning', message);
  }

  return {
    showError,
    showSuccess,
    showInfo,
    showWarning,
  };
}
