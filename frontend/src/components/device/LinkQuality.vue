<script setup lang="ts">
  import { getSignalIcon } from '@/modules/formatters/device.formatter';
  import Icon from '../controls/Icon.vue';

  const props = defineProps({
    value: Number,
  });
</script>
<template>
  <span
    key="linkquality"
    className="me-"
    v-if="props.value !== undefined">
    <Icon :icon="getSignalIcon(props.value)" />
    {{ props.value + ' LQI' }}
  </span>
</template>
