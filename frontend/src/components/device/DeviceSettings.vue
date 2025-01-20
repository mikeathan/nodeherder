<script setup lang="ts">
import { store } from '../../store/index';
import { computed } from 'vue';
import Toggle from '../input/Toggle.vue';
import InputBox from '../input/InputBox.vue';
import { DeviceSettings, TimeInterval } from '@/types/settings.type';

const props = defineProps({
  id: { type: String, required: true },
});

const deviceSettings = computed(() => {
  return store.getters['hub/findDeviceSetting'](props.id);
});

function toggleChanged(propName: any, propValue: any) {
  save(propName, propValue);
}

function inputLostFocus(propName: any, propValue: any) {
  save(propName, propValue);
}
function inputTimeIntervalLostFocus(propName: any, propValue: any) {
  const timeInterval = deviceSettings.value[propName] as TimeInterval;
  timeInterval.value = propValue;
  save(propName, timeInterval);
}
function save(propName: any, propValue: any) {
  if (deviceSettings.value[propName] != propValue) {
    deviceSettings.value[propName] = propValue;
    store.dispatch('hub/saveDeviceSettings', deviceSettings.value as DeviceSettings);
  }
}
function isTimeInterval(value: any): value is TimeInterval {
  return typeof value === 'object' && value !== null && 'value' in value && 'unit' in value;
}
</script>

<template>
  <div class="grid col-12 align-items-center grid-nogutter" v-for="(value, key) in deviceSettings" :key="key">
    <dl class="col-12 md:col-3">
      <dt class="text-secondary">
        <strong> {{ key }}</strong>
      </dt>
    </dl>
    <div class="md:col-4">
      <div v-if="typeof value === 'boolean'">
        <Toggle :minimal="false" :value="value" :valueOn="true" :value-off="false"
          @update="(v) => toggleChanged(key, v)">
        </Toggle>
      </div>
      <div v-else-if="isTimeInterval(value)">
        <InputBox :label="value.unit" :value="value.value" :is-numeric="true"
          @lost-focus="(f) => inputTimeIntervalLostFocus(key, f)" />
      </div>
      <div v-else>
        <InputBox :value="value" :disabled="typeof value !== 'number'" :is-numeric="typeof value === 'number'"
          @lost-focus="(f) => inputLostFocus(key, f)">
        </InputBox>
      </div>
    </div>
  </div>
</template>
