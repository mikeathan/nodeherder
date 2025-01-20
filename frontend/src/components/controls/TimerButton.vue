<script setup lang="ts">
  import { store } from '@/store';
  import { BridgeSettingsType } from '@/types/settings.type.type';
  import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
  import { computed } from 'vue';

  const bridgeConfig = computed<BridgeSettingsType>(() => {
    return store.getters['hub/bridge']() as BridgeSettingsType;
  });

  function enablePermitJoin() {
    store.dispatch('hub/enablePermitJoin', props.duration);
  }

  function disablePermitJoin() {
    store.dispatch('hub/disablePermitJoin');
  }

  const props = defineProps({
    duration: { type: Number, default: 60 },
  });

  let intervalId: any = null;
  const remainingTime = ref(
    localStorage.getItem('remainingTime') ? parseInt(localStorage.getItem('remainingTime')!) : props.duration
  );

  const isRunning = ref(false);
  const isDisabled = ref(false);

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
      } else {
        stopTimer();
      }
    },
    { deep: true }
  );

  const buttonLabel = computed(() => {
    if (isRunning.value) {
      return `Disable Join [${formattedTime.value}]`;
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
  .toggle-btn {
    text-decoration: none;
    color: inherit;
    cursor: pointer;
    background-color: transparent;
    color: var(--primary-color);

    display: inline-flex;
    align-items: center;
  }

  .toggle-btn .small-text {
    font-size: 0.9rem;
  }

  .toggle-btn .toggle-icon {
    margin-right: 0.5rem;
  }

  .toggle-btn.running {
    color: var(--bs-danger-border-subtle);
    font-weight: 500;
  }

  .toggle-btn.disabled {
    cursor: default;
    color: var(--bs-border-color);
    pointer-events: none;
  }
  .toggle-btn.p-button:hover {
    background-color: transparent;
    border-color: transparent;
    color: var(--primary-color);
  }
</style>
<template>
  <a
    class="toggle-btn"
    :class="{ running: isRunning, disabled: isDisabled }"
    href="#"
    @click.prevent="isDisabled ? null : toggleTimer()">
    <span class="toggle-icon pi pi-sitemap"></span>
    <span class="small-text">{{ buttonLabel }}</span>
  </a>
</template>
