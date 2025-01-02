<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from "vue";
import ProgressBar from "primevue/progressbar";
import { computed } from "vue";

const totalTime = 60;
const progressValue = ref(0);
const timeLeft = ref(totalTime);
let intervalId: any = null;

const buttonLabel = computed(() => {
    if (timeLeft.value > 0) {
        return "Permit Join";
    }
    return "Disable Join";
});

const isRunning = ref<boolean>(false);
const formattedTime = computed(() => {
    const minutes = Math.floor(timeLeft.value / 60);
    const seconds = timeLeft.value % 60;
    return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
});

const startTimer = () => {
    isRunning.value = true;
    intervalId = setInterval(() => {
        if (timeLeft.value > 0) {
            timeLeft.value -= 1;
            progressValue.value = ((totalTime - timeLeft.value) / totalTime) * 100;
        } else {
            clearTimer()
        }
    }, 1000);
};

const clearTimer = () => {
    clearInterval(intervalId);
    intervalId = null;
    timeLeft.value = totalTime;
    progressValue.value = 0;
    isRunning.value = false;
}
const stopTimer = () => {
    clearTimer()
};

const toggleTimer = () => {
    if (isRunning.value) {
        stopTimer();
    } else {
        startTimer();
    }
};

onBeforeUnmount(() => {
    clearTimer();// ?? not sure  need that as i dont want to clear ifi refresh ?/
});

</script>
<style scoped>
.timer-panel {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
    background-color: #f8f9fa;
    width: auto;
}

.content {
    display: flex;
    align-items: center;
    gap: 8px;
}

.time-label {
    font-size: 14px;
    font-weight: bold;
    margin: 0;
    color: #333;
}

.progress-container {
    position: relative;
    width: 120px;
    /* Adjust width as needed */
    height: 16px;
    /* Match the height of the progress bar */
}

.progress-bar {
    width: 100%;
    height: 100%;
}

.progress-text {
    position: absolute;
    top: 0;
    left: 50%;
    transform: translate(-50%, 0);
    font-size: 12px;
    font-weight: bold;
    color: #000;
    /* Adjust color to match your theme */
    line-height: 16px;
    /* Match the height of the progress bar */
    text-align: center;
    pointer-events: none;
    /* Ensure text is not clickable */
}

.start-btn {
    background-color: #28a745;
    /* Green for start */
    color: white;
    border: none;
}

.stop-btn {
    background-color: #dc3545;
    /* Red for stop */
    color: white;
    border: none;
}

button {
    font-size: 14px;
    padding: 4px 8px;
}

.no-text .p-progressbar-value {
    color: transparent !important;
    /* Hide default progress value */
}
</style>
<template>
    <div class="timer-panel">
        <div class="content">
            <p class="time-label">{{ formattedTime }}</p>
            <div class="progress-container">
                <ProgressBar :value="progressValue" class="progress-bar no-text" />
                <div class="progress-text">{{ progressValue.toFixed(0) }}%</div>
            </div>
            <Button :label="isRunning ? 'Stop' : 'Start'" :class="{ 'start-btn': !isRunning, 'stop-btn': isRunning }"
                @click="toggleTimer" size="small" />
        </div>
    </div>
</template>