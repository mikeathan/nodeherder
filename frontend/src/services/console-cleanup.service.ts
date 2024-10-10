import { StoreType } from '@/store/index';
import { onUnmounted, ref } from 'vue';

const EXPIRATION_TIMEOUT = 5 * 60 * 1000; // 5 minutes

const isTimerRunning = ref(false);
const isComponentMounted = ref(false);

export const consoleCleanupService = {
  startTimer(store: StoreType) {
    isComponentMounted.value = true;

    if (!isTimerRunning.value) {
      const intervalId = setInterval(() => {
        // if component is unmounted and there are no messages in the store, stop the timer
        // else keep the timer running
        if (
          !isComponentMounted.value &&
          store.getters['console/messages']().length === 0
        ) {
          clearInterval(intervalId);

          isTimerRunning.value = false;
          console.log('console-cleanup: stopped');
          return;
        }

        store.commit('console/removeExpiredMessages');
      }, EXPIRATION_TIMEOUT);

      isTimerRunning.value = true;
      console.log('console-cleanup: started');
    }

    onUnmounted(() => {
      isComponentMounted.value = false;
    });
  },
};
