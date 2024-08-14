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
  if (!store.getters['appconfig/initialized']() as Boolean) {
    store.dispatch('ws/emit', { event: 'loadAppConfig' });
  }

  return store.getters['appconfig/findDeviceSetting'](props.id)
});


function toggleChanged(propName: any, propValue: any) {
  save(propName, propValue);
}

function inputLostFocus(propName: any, propValue: any) {
  save(propName, propValue);
}

function save(propName: any, propValue: any) {
  if (deviceSettings.value[propName] != propValue) {
    deviceSettings.value[propName] = propValue;
    store.dispatch('appconfig/saveDeviceSettings', deviceSettings.value as DeviceSettings);
  }
}

</script>
<template>
  <div class="row border-bottom py-1 w-100 align-items-center" v-for="(value, key) in deviceSettings" :key="key">
    <dl class="col-12 col-md-3">
      <dt>
        <strong> {{ key }}</strong>
      </dt>
    </dl>
    <div class="col-md-4">
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
