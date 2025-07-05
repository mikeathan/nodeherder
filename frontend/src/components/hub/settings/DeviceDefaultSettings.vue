<script setup lang="ts">
  import { computed } from 'vue';
  import { store } from '../../../store/index';
  import { DeviceConfig } from '@/types/settings.type';
  import InputBox from '../../input/InputBox.vue';
  import Toggle from '@/components/input/Toggle.vue';
  import { DeviceConfigDefaultComponents } from '@/mixins/useSettingsComponents';

  const deviceDefaultSettings = computed(() => {
    return store.getters['hub/deviceDefaults']() as DeviceConfig;
  });

  function inputLostFocus(propName: any, propValue: any) {
    save(propName, propValue);
  }

  function save<K extends keyof DeviceConfig>(propValue: DeviceConfig[K], key: K) {

  //   if (!deviceSettings.value) return;

  //   if (!localOverride.value) {
  //     localOverride.value = { ...deviceSettings.value };
  //   }

  //   if (!isObject(localOverride.value[propName])) {
  //     if (localOverride.value[propName] !== propValue) {
  //       (localOverride.value as any)[propName] = propValue;
  //     }
  //   }

  //   store.dispatch('hub/saveDeviceConfigOverrides', localOverride.value as DeviceConfig);
  // }
    if (deviceDefaultSettings.value[key] != propValue) {
      deviceDefaultSettings.value[key] = propValue;

      store.dispatch('hub/saveDeviceConfigDefaults', deviceDefaultSettings.value);
    }
  }
  function inputUpdated(propName: keyof DeviceConfig, propValue: any) {
    save(propName, propValue);
  }
  function updateState<K extends keyof DeviceConfig>(enabled: DeviceConfig[K], key: K) {
    save(enabled, key);
  }
  function isObject(value: any): value is object {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
  }
</script>

<template>
  <h3>Device Defaults</h3>
  <div class="pt-3" />
  <div class="grid grid-nogutter" v-for="(value, key) in deviceDefaultSettings" :key="key">
    <dl class="col-12 md:col-3 text-secondary">
      <dt>
        <strong>{{ key }}</strong>
      </dt>
    </dl>
    <div class="col-12 md:col-3 pb-3">
      <div v-if="DeviceConfigDefaultComponents[key]">
        <component
          :is="DeviceConfigDefaultComponents[key]"
          :value="value"
          @update="(v: any) => inputUpdated(key as keyof DeviceConfig, v)" />
      </div>
      <template v-if="typeof value === 'boolean'">
        <Toggle :value="value" :valueOn="true" :valueOff="false" @update="(v: any) => updateState(v, key)" />
      </template>
      <template v-else-if="typeof value === 'number' || typeof value === 'string'">
        <InputBox :value="value" :is-numeric="typeof value == 'number'" @lost-focus="(f) => inputLostFocus(key, f)" />
      </template>
    </div>
  </div>
</template>
