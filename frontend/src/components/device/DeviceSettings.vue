<script setup lang="ts">
import { store } from '../../store/index';
import { computed } from 'vue';
import Toggle from '../input/Toggle.vue';
import { DeviceSettings } from '@/types/settings';
import InputBox from '../input/InputBox.vue';

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
  console.log('inputLostFocus', propName, propValue);
  save(propName, propValue);
}

function save(propName: any, propValue: any) {
  if (deviceSettings.value[propName] != propValue) {
    deviceSettings.value[propName] = propValue;
    store.dispatch('hub/saveDeviceSettings', deviceSettings.value as DeviceSettings);
  }
}
</script>

to fix the duration , give it vlaue as it doesnt make sense when tis 500000
<template>
  <div class="grid col-12 align-items-center grid-nogutter" v-for="(value, key) in deviceSettings" :key="key">
    <dl class="col-12 md:col-3">
      <dt class="text-secondary">
        <strong> {{ key }}</strong>
      </dt>
    </dl>
    <div class="md:col-4">
      <div v-if="typeof value === 'boolean'">
        <Toggle :minimal="false" :value="value" :valueOn="true" :valueoff="false"
          @update="(v) => toggleChanged(key, v)">
        </Toggle>
      </div>
      <div v-else>
        <InputBox :value="value" :disabled="typeof value !== 'number'" :is-numeric="typeof value === 'number'"
          @lost-focus="(f) => inputLostFocus(key, f)">
        </InputBox>
      </div>
    </div>
  </div>
</template>
