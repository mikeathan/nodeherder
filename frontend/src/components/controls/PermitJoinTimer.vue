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
    padding: 0.5rem;
    border-radius: 8px;

    background-color: #2c2f36;
    color: #e0e0e0;
    border: 1px solid #41444b;
    text-decoration: none;

    font-size: 1rem;
    font-weight: 500;

    transition: background-color 0.3s, border-color 0.3s;
    gap: 0.75rem;
  }

  .join-control:hover {
    background-color: #34383f;
    border-color: #555960;
  }

  .join-control.running {
    background-color: #37424b;
    border-color: #5a6670;
  }

  .join-control.disabled {
    opacity: 0.6;
    pointer-events: none;
  }

  .join-label {
    font-size: 1rem;
  }

  .join-timer {
    font-family: monospace;
    background-color: #1f252b;
    padding: 0.2rem 0.75rem;
    border-radius: 4px;
    border: 1px solid #3b3f46;
    color: #cbd5e1;
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
