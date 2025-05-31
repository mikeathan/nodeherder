<script setup lang="ts">
  import { store } from '@/store';
  import { BridgeSettingsType } from '@/types/settings.type';
  import { ref, onMounted, watch, watchEffect } from 'vue';
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
  const emit = defineEmits<{
    (e: 'statusUpdated', value: boolean): void;
  }>();

  const props = defineProps({
    duration: { type: Number, default: 60 },
    allowJoin: { type: Boolean, default: false },
  });

  let intervalId: any = null;
  const remainingTime = ref(
    localStorage.getItem('remainingTime') ? parseInt(localStorage.getItem('remainingTime')!) : props.duration
  );

  const isRunning = ref(false);
  const isDisabled = ref(false);

  watch(
    () => props.allowJoin,
    (newValue) => {

      need to guard agains same status changes
      console.log('watch allowJoin:', props.allowJoin, ' newValue:', newValue);

      if (newValue) {
        enablePermitJoin();
        console.log('enablePermitJoin');
      } else {
        disablePermitJoin();
        console.log('disablePermitJoin');
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
      emit('statusUpdated',  isRunning.value); ???
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
      console.log('disablePermitJoin');
    } else {
      enablePermitJoin();
      console.log('enablePermitJoin');
    }
    isDisabled.value = true;
  };

  onMounted(() => {
    console.log('onMounted isRunning:', isRunning.value);
    if (isRunning.value) {
      startTimer();
    }
  });
</script>

<style scoped>
  .content {
    display: flex;
    align-items: center;
    width: 100%;
    background: var(--p-content-background);
    color: var(--p-content-color);
    border: 1px solid var(--bs-border-color);
    padding: var(--p-list-option-padding);
    border-radius: var(--p-border-radius-md);
  }
</style>
<template>
  <a
    v-if="isRunning"
    class="content toggle-btn"
    :class="{ running: isRunning, disabled: isDisabled }"
    href="#"
    @click.prevent="isDisabled ? null : toggleTimer()">
    <span class="small-text">{{ buttonLabel }}</span>
  </a>
</template>
