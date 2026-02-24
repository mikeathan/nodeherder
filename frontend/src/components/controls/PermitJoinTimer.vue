<script setup lang="ts">
  import { store } from '@/store';
  import { BridgeSettingsType } from '@/types/settings.type';
  import { ref, onMounted, watch } from 'vue';
  import { computed } from 'vue';

  const emit = defineEmits<{
    (e: 'statusUpdated', value: boolean): void;
  }>();

  const props = defineProps({
    duration: { type: Number, default: 60 },
    allowJoin: { type: Boolean, default: false },
  });

  const bridgeConfig = computed<BridgeSettingsType>(() => {
    return store.getters['hub/bridge']() as BridgeSettingsType;
  });

  function enablePermitJoin() {
    store.dispatch('hub/enablePermitJoin', props.duration);
  }

  function disablePermitJoin() {
    store.dispatch('hub/disablePermitJoin');
  }

  let intervalId: any = null;
  const remainingTime = ref(
    localStorage.getItem('remainingTime') ? parseInt(localStorage.getItem('remainingTime')!) : props.duration
  );

  const isRunning = ref(false);
  const isDisabled = ref(false);

  watch(
    () => props.allowJoin,
    (newValue) => {
      if (newValue && !isRunning.value) {
        enablePermitJoin();
      } else if (!newValue && isRunning.value) {
        disablePermitJoin();
      }
    }
  );

  watch(
    bridgeConfig,
    (newValue, oldValue) => {
      if (newValue.permitJoin == isRunning.value) {
        return;
      }
      isDisabled.value = false;
      isRunning.value = newValue.permitJoin;

      if (newValue.permitJoin) {
        startTimer();

        emit('statusUpdated', isRunning.value);
      } else {
        stopTimer();
      }
    },
    { deep: true }
  );

  const buttonLabel = computed(() => {
    if (isRunning.value) {
      return `Disable Join`;
    }
    return 'Permit Join';
  });

  const formattedTime = computed(() => {
    const minutes = Math.floor(remainingTime.value / 60);
    const seconds = remainingTime.value % 60;

    return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
  });

  const startTimer = () => {
    intervalId = setInterval(() => {
      if (isDisabled.value) {
        return;
      }
      if (remainingTime.value > 0) {
        remainingTime.value -= 1;
        localStorage.setItem('remainingTime', remainingTime.value.toString());
      } else {
        stopTimer();
      }
    }, 1000);
  };

  const stopTimer = () => {
    clearInterval(intervalId);
    intervalId = null;
    setRemainingTime(props.duration);

    emit('statusUpdated', isRunning.value);
  };

  const setRemainingTime = (value: number) => {
    remainingTime.value = value;
    localStorage.setItem('remainingTime', remainingTime.value.toString());
  };

  const toggleTimer = () => {
    if (isRunning.value) {
      disablePermitJoin();
    } else {
      enablePermitJoin();
    }
    isDisabled.value = true;
  };

  onMounted(() => {
    if (isRunning.value) {
      startTimer();
    }
  });
</script>

<style scoped>
  .join-control {
    display: flex;
    align-items: center;
    justify-content: center;

    width: 100%;
    padding: 0.5rem 1rem;
    border-radius: 0;
    /* removed border radius for banner-like appearance if it sits at the top, or 8px if inside a card */

    background-color: var(--p-surface-800);
    color: var(--p-surface-0);
    border-bottom: 1px solid var(--p-surface-700);
    text-decoration: none;

    font-size: 0.875rem;
    font-weight: 500;

    transition:
      background-color 0.2s ease,
      border-color 0.2s ease;
    gap: 0.75rem;
  }

  .join-control:hover {
    background-color: var(--p-surface-700);
  }

  .join-control.running {
    background-color: var(--p-surface-800);
  }

  .join-control.disabled {
    opacity: 0.5;
    pointer-events: none;
  }

  .join-label {
    font-size: 0.875rem;
    font-weight: 600;
  }

  .join-timer {
    font-family: inherit;
    font-size: 0.875rem;
    font-weight: 700;
    background-color: var(--p-surface-900);
    padding: 0.2rem 0.6rem;
    border-radius: 6px;
    border: 1px solid var(--p-surface-700);
    color: var(--p-surface-0);
    letter-spacing: 0.05em;
  }

  /* ── Light theme overrides (Targeting Tailwind's dark mode absence) ── */
  :global(html:not(.dark)) .join-control,
  :global(html:not(.dark)) .join-control.running {
    background-color: var(--p-surface-100);
    color: var(--p-surface-900);
    border-bottom-color: var(--p-surface-200);
  }

  :global(html:not(.dark)) .join-control:hover,
  :global(html:not(.dark)) .join-control.running:hover {
    background-color: var(--p-surface-200);
  }

  :global(html:not(.dark)) .join-timer {
    background-color: var(--p-surface-0);
    color: var(--p-primary-600, #2563eb);
    border-color: var(--p-surface-300);
  }
</style>

<template>
  <a
    v-if="isRunning"
    class="join-control running"
    :class="{ disabled: isDisabled }"
    href="#"
    @click.prevent="isDisabled ? null : toggleTimer()">
    <span class="join-label">{{ buttonLabel }}</span>
    <span class="join-timer">{{ formattedTime }}</span>
  </a>
</template>
