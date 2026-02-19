<script setup lang="ts">
  import { getSignalIcon } from '@/modules/formatters/device.formatter';
  import Icon from '../controls/Icon.vue';

  const props = defineProps({
    value: Number,
  });
</script>
<template>
  <div key="linkquality" class="flex align-items-center mr-2" v-if="props.value">
    <Icon :icon="getSignalIcon(props.value)" />
    <span class="ml-1">{{ props.value + ' LQI' }}</span>
  </div>
</template>
