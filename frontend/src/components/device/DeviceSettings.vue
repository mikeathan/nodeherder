<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref } from 'vue';
import Toggle from '../input/Toggle.vue';
import { DeviceSettings } from '@/types/settings';
import { createDeviceSettings } from '@/contracts/settings';
import InputBox from '../input/InputBox.vue';
import { KeyyValuePair } from '@/types/types';

const props = defineProps({
  id: { type: String, required: true },
});
const cachedDeviceSettings = ref<KeyyValuePair<any>>(
  {} as KeyyValuePair<any>,
);
const isDirty = computed(() => {
  return (
    JSON.stringify(cachedDeviceSettings.value) !==
    JSON.stringify(deviceSettings.value)
  );
});

const deviceSettings = computed(() => {
  if (
    !store.getters['appconfig/initialized']() as Boolean
  ) {
    store.dispatch('ws/emit', { event: 'loadAppConfig' });
  }

  const settings = store.getters[
    'appconfig/findDeviceSetting'
  ](props.id);

  if (!settings) {
    const newDeviceSettings = JSON.parse(
      JSON.stringify(createDeviceSettings(props.id)),
    );
    store.dispatch(
      'appconfig/saveDeviceSettings',
      newDeviceSettings,
    );
    return newDeviceSettings;
  }
  return JSON.parse(JSON.stringify(settings));
});

onMounted(() => {
  cachedDeviceSettings.value = JSON.parse(
    JSON.stringify(deviceSettings.value),
  );
});

function updateValue(propName: any, propValue: any) {
  cachedDeviceSettings.value[propName] = propValue;
}

function save() {
  store.dispatch(
    'appconfig/saveDeviceSettings',
    cachedDeviceSettings.value as DeviceSettings,
  );
}
</script>
<template>
  <div
    class="row border-bottom py-1 w-100 align-items-center"
    v-for="(value, key) in deviceSettings"
    :key="key">
    <dl class="col-12 col-md-3">
      <dt>
        <strong> {{ key }}</strong>
      </dt>
    </dl>
    <div class="col-md-4">
      <div v-if="typeof value === 'boolean'">
        <Toggle
          :minimal="false"
          :value="value"
          :valueOn="true"
          :valueoff="false"
          @update="(v) => updateValue(key, v)">
        </Toggle>
      </div>
      <div v-else>
        <InputBox
          @updated="(v) => updateValue(key, v)"
          :value="value"
          :disabled="typeof value !== 'number'"
          :is-numeric="typeof value === 'number'">
        </InputBox>
      </div>
    </div>
  </div>
  <br />
  <div class="pb-3">
    <button
      type="button"
      class="btn btn-light"
      aria-label="Save"
      @click="save()"
      :disabled="!isDirty">
      Save
    </button>
  </div>
</template>
