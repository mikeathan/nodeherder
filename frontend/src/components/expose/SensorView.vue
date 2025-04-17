<script setup lang="ts">
import { ref, computed } from 'vue';
import Brightness from '../expose/brightness.vue';

import {
  mdiCeilingLightMultiple
} from '@mdi/js';

const props = defineProps({
  value: {
    type: Number,
    default: 50,
  },
});
const value = ref(props.value);

function updateValue(newValue: number): void {
  value.value = newValue;
}
</script>

<style scoped></style>

<template>
  <Card>
    <template #title>
      <Icon :icon="{ name: mdiCeilingLightMultiple, color: 'yellow' }" width="32" height="32" />
      Brightness
      {{ value }}
    </template>
    <template #content>
      <div>
        <Brightness :value="value" @update="updateValue" />
      </div>
    </template>
  </Card>
</template>
