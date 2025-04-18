<script setup lang="ts">
import { ref, computed } from 'vue';
import Brightness from '../expose/brightness.vue';

import { mdiCeilingLightMultiple } from '@mdi/js';

const props = defineProps({
  value: {
    type: Number,
    default: 50,
  },
});
const value = ref(props.value);
const lightOn = ref(true);

function updateValue(newValue: number): void {
  value.value = newValue;
}
function handleIconClick(): void {
  lightOn.value = !lightOn.value;
}
</script>

<style scoped>
.entity-card {
  padding: 1em;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  gap: 1em;
}

.entity-header {
  display: flex;
  align-items: center;
  gap: 0.75em;
}

.entity-labels {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.entity-labels {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.entity-title {
  font-size: 1rem;
  font-weight: 500;
}

.entity-value {
  font-size: 0.75rem;
  color: #777;
  margin-top: 0.2em;
}

.entity-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #fff8e1;
  cursor: pointer;
  transition: background-color 0.2s;
}
</style>

<template>
  when off
  make slider grey and value should say off

  ## Sensor is Binary ##
  allows write ( eg switch, alarm)
  we can toggle the icon
  icon gets gray when is off
  the value is set to off

  allows only read (eg presence sensor)
  icon is grey when is off,
  we cant click the icon
  the value is set to off


  ## Sensor is Number ##
  allows read only (eg temperature, humidity, battery)
  Icon is same as curren device card
  layout is same as name of sensor and below value with unit

  allows write (eg light)
  we can toggle the icon and the value when off displays off
  when is off the slider is gray and disabled

  think how to deal with enums and how to group them

  ## Sensor is Enum ##




  <Card>
    <template #title>
      <div class="entity-header">
        <div class="entity-icon" @click="handleIconClick">
          <Icon :icon="{ name: mdiCeilingLightMultiple, color: lightOn ? '#ffc107' : '#9e9e9e' }" width="28"
            height="28" />
        </div>
        <div class="entity-labels">
          <div class="entity-title">Brightness</div>
          <div class="entity-value">{{ value }}%</div>
        </div>
      </div>
    </template>
    <template #content>
      <div>
        <Brightness :value="value" @update="updateValue" :min="0" :max="100" />
      </div>
    </template>
  </Card>
</template>
