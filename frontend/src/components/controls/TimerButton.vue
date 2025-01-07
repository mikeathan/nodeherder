<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { computed } from 'vue';

const props = defineProps({
  duration: { type: Number, default: 60 },
  startEvent: {
    type: Function,
    default: () => { },
  },
  stopEvent: {
    type: Function,
    default: () => { },
  },
});

let intervalId: any = null;
const remainingTime = ref(
  localStorage.getItem('remainingTime') ? parseInt(localStorage.getItem('remainingTime')!) : props.duration
);

const isRunning = ref(localStorage.getItem('isRunning') === 'true' ? true : false);
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
  setIsRunning(false);
  setRemainingTime(props.duration);
};

const setIsRunning = (value: boolean) => {
  isRunning.value = value;
  localStorage.setItem('isRunning', isRunning.value.toString());
};

const setRemainingTime = (value: number) => {
  remainingTime.value = value;
  localStorage.setItem('remainingTime', remainingTime.value.toString());
};
const toggleTimer = () => {
  isRunning.value = !isRunning.value;
  localStorage.setItem('isRunning', isRunning.value.toString());

  if (isRunning.value) {
    onStart();
  } else {
    onStopped();
  }
};

const onStart = () => {
  props.startEvent();
  startTimer();
};

const onStopped = () => {
  props.stopEvent();
  stopTimer();
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
}

.toggle-btn.p-button:hover {
  background-color: transparent;
  border-color: transparent;
  color: var(--primary-color);
}

.toggle-btn .small-text {
  font-size: 0.9rem;
  /* Adjust the value as needed */
}

.toggle-btn .toggle-icon {
  margin-right: 0.5rem;
}
</style>
<template>
  <a class="toggle-btn" href="javascript:void(0)" v-on:click="toggleTimer()">
    <span class="toggle-icon pi pi-sitemap"></span>
    <span class="small-text">{{ buttonLabel }}</span>
  </a>
</template>
