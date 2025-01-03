<script setup lang="ts">
  import { ref, onMounted, onBeforeUnmount } from 'vue';
  import ProgressBar from 'primevue/progressbar';
  import { computed } from 'vue';

  const totalTime = 60;
  const remainingTime = ref(
    localStorage.getItem('remainingTime') ? parseInt(localStorage.getItem('remainingTime')!) : totalTime
  );
  let intervalId: any = null;

  const buttonLabel = computed(() => {
    if (isRunning.value) {
      return 'Disable Join';
    }
    return 'Permit Join';
  });

  const isRunning = ref(localStorage.getItem('isRunning') === 'true' ? true : false);
  const formattedTime = computed(() => {
    const minutes = Math.floor(remainingTime.value / 60);
    const seconds = remainingTime.value % 60;

    return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
  });

  const startTimer = () => {
    console.log('Starting timer');
    intervalId = setInterval(() => {
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
    isRunning.value = false;
    localStorage.setItem('isRunning', isRunning.value.toString());
    remainingTime.value = totalTime;
    localStorage.setItem('remainingTime', remainingTime.value.toString());
  };

  const toggleTimer = () => {
    isRunning.value = !isRunning.value;
    localStorage.setItem('isRunning', isRunning.value.toString());

    if (isRunning.value) {
      startTimer();
    } else {
      stopTimer();
    }
  };

  onMounted(() => {
    if (isRunning.value) {
      startTimer();
    }
  });
</script>

<style scoped>
  .content {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .time-label {
    font-size: 14px;
    margin: 0;
  }
</style>
<template>
  <div class="content">
    <p class="time-label" v-if="isRunning">{{ formattedTime }}</p>
    <Button
      text
      severity="secondary"
      :label="buttonLabel"
      :class="{
        'start-btn': !isRunning,
        'stop-btn': isRunning,
      }"
      :icon="isRunning ? 'pi pi-pause' : 'pi pi-play'"
      @click="toggleTimer"
      size="small" />
  </div>
</template>
