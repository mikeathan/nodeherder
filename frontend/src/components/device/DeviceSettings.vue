<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed } from 'vue';
  import Toggle from '../input/Toggle.vue';
  import InputBox from '../input/InputBox.vue';
  import { DeviceConfig } from '@/types/settings.type';
  import { ExposeSettingsComponents } from '@/mixins/useSettingsComponents';

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

  // TODO: needs refactoring to emit only if sth has changed but i have problesm with new debouncer map prop
  function save(propName: string, propValue: any) {
    if (!isObject(deviceSettings.value[propName])) {
      if (deviceSettings.value[propName] != propValue) {
        deviceSettings.value[propName] = propValue;
      }
    }

    store.dispatch('hub/saveDeviceSettings', deviceSettings.value as DeviceConfig);
  }

  function isObject(value: any): value is object {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
  }
  // TODO: needs refactoring to emit only if sth has changed but i have problesm with new debouncer map prop

  function inputUpdated(propName: any, propValue: any) {
    save(propName, { ...propValue });
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
      <div v-if="ExposeSettingsComponents[key]">
        <component
          :is="ExposeSettingsComponents[key]"
          v-bind="{
            id: props.id,
            value: value,
          }"
          @update="(v:any) => inputUpdated(key, v)" />
      </div>
      <div v-else-if="typeof value === 'boolean'">
        <Toggle :value="value" :valueOn="true" :valueOff="false" @update="(v) => toggleChanged(key, v)"> </Toggle>
      </div>
      <div v-else>
        <InputBox
          :value="value"
          :disabled="typeof value !== 'number'"
          :is-numeric="typeof value === 'number'"
          @lost-focus="(f) => inputLostFocus(key, f)">
        </InputBox>
      </div>
    </div>
  </div>
</template>
