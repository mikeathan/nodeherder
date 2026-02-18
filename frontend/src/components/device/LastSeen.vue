<script setup>
  import { watch, ref, onMounted, onUnmounted } from 'vue';
  import ElapsedTimer from '../../modules/time-elapsed';

  const props = defineProps({
    timestamp: String,
    truncated: {
      type: Boolean,
      default: true,
    },
  });

  let elapsedTimer = null;
  let lastSeenElement = ref(null);

  watch(
    () => props.timestamp,
    (newlastSeen) => {
      if (lastSeenElement.value == undefined) {
        return;
      }

      elapsedTimer.Format(newlastSeen);
    },
    { immediate: true }
  );

  onMounted(() => {
    elapsedTimer = new ElapsedTimer(lastSeenElement.value);
    elapsedTimer.Format(props.timestamp);
  });

  onUnmounted(() => {
    elapsedTimer.dispose();
    // DO I NEED THAT
    // elapsedTimer = null; // Clear reference
  });
</script>
<template>
  <div
    :title="'last update ' + timestamp"
    :ref="
      (el) => {
        lastSeenElement = el;
      }
    "
    :class="{ 'col white-space-nowrap overflow-hidden text-overflow-ellipsis': truncated }"></div>
</template>
