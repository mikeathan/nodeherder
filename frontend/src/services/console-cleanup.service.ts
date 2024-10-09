import { StoreType } from '@/store/index';
import { onUnmounted, ref } from 'vue';

const EXPIRATION_TIMEOUT = 30 * 1000; // 30  sec

const isTimerRunning = ref(false);
const isComponentMounted = ref(false);


TODO - fix the timer dipsoe when unmoun
export const consoleCleanupService = {
  startTimer(store: StoreType) {
    console.log('console-cleanup: mounted');

    if (!isTimerRunning.value) {
      isComponentMounted.value = true;

      console.log('console-cleanup: started');

      const intervalId = setInterval(() => {
        console.log(
          'console-cleanup: deleteExpiredMessages',
        );

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

      console.log(
        'console-cleanup: intervalId',
        intervalId,
        ' timeisrunning ',
        isTimerRunning.value,
      );
    }

    onUnmounted(() => {
      console.log('console-cleanup: unmounted');

      isComponentMounted.value = true;
    });
  },
};
