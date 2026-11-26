/*
 * Connection notifications are produced from transport transitions, not from components.
 * One subscription owns dedupe and cleanup, so remounting a page cannot duplicate history
 * or leak a listener.
 */
import { onBeforeUnmount } from 'vue';
import { store } from '@/store';
import type { ConnectionStatusType } from '@/types/connection.type';
import { connectionTransition } from '@/domain/notifications';

export function useNotificationProducers() {
  let previous = store.getters['ws/getConnectionStatus'] as ConnectionStatusType;
  let observedLoss = false;

  const unsubscribe = store.subscribe((mutation) => {
    if (mutation.type !== 'ws/setConnectionStatus') return;
    const next = mutation.payload as ConnectionStatusType;
    if (next === previous) return;

    const transition = connectionTransition(previous, next);
    previous = next;
    if (transition === 'lost') {
      observedLoss = true;
      void store.dispatch('notifications/notify', {
        severity: 'warn',
        category: 'connection',
        title: 'Hub connection lost',
        body: 'Live updates are paused. Values show the last snapshot until the link returns.',
        dedupeKey: 'hub-connection-lost',
        action: { label: 'Open logs', to: '/consoleviewer' },
      });
      return;
    }
    if (transition === 'restored' && observedLoss) {
      observedLoss = false;
      void store.dispatch('notifications/notify', {
        severity: 'success',
        category: 'connection',
        title: 'Hub connection restored',
        body: 'Live updates resumed. Values that changed while you were away are now current.',
        dedupeKey: 'hub-connection-restored',
      });
    }
  });

  onBeforeUnmount(unsubscribe);
}
