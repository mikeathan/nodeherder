/*
 * Single shared open state for the notification inbox. The header bell owns the dialog;
 * the Overview digest only asks it to open, so there is no second copy of the inbox.
 */
import { ref } from 'vue';

const open = ref(false);

export function useNotificationPanel() {
  return {
    open,
    show: () => (open.value = true),
    hide: () => (open.value = false),
  };
}
