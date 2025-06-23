<script setup lang="ts">
  import { computed } from 'vue';
  import { store } from '../../../store/index';
  import { HistorySettingsType, HistorySettingsPropsType, DeviceConfig } from '@/types/settings.type';
  import InputBox from '../../input/InputBox.vue';
  import Toggle from '@/components/input/Toggle.vue';
  import { isTimeInterval, isDeviceDebounce } from '@/contracts/settings';
  import DebounceSettings from '@/components/settings/DebounceSettings.vue';

  const deviceDefaultSettings = computed(() => {
    return store.getters['hub/deviceDefaults']() as DeviceConfig;
  });

  function inputLostFocus(propName: any, propValue: any) {
    console.log('inputLostFocus', propName, propValue);
    // save(propName, propValue);
  }

  // function save(propName: HistorySettingsPropsType, propValue: any) {
  //   if (deviceDefaultSettings.value[propName].value != propValue) {
  //     deviceDefaultSettings.value[propName].value = propValue;
  //     store.dispatch('hub/saveHistorySettings', deviceDefaultSettings.value);
  //   }
  // }
  function updateState<K extends keyof DeviceConfig>(enabled: DeviceConfig[K], key: K) {
    if (enabled === deviceDefaultSettings.value[key]) return;

    deviceDefaultSettings.value[key] = enabled;
  }

  TODO;
</script>

<template>
  <h3>Device</h3>
  <div class="pt-3" />
  <div class="grid grid-nogutter" v-for="(value, key) in deviceDefaultSettings" :key="key">
    <dl class="col-12 md:col-3 text-secondary">
      <dt>
        <strong>{{ key }}</strong>
      </dt>
    </dl>
    <div class="col-12 md:col-3 pb-3">
      <template v-if="typeof value === 'boolean'">
        <Toggle :value="value" :valueOn="true" :valueOff="false" @update="(v: any) => updateState(v, key)" />
      </template>
      <template v-else-if="typeof value === 'number' || typeof value === 'string'">
        <InputBox :value="value" :is-numeric="typeof value == 'number'" @lost-focus="(f) => inputLostFocus(key, f)" />
      </template>
      <template v-else-if="isTimeInterval(value)">
        <InputBox
          :label="value.unit"
          :value="value.value"
          :is-numeric="true"
          @lost-focus="(f) => inputLostFocus(key, f)" />
      </template>
      <template v-else-if="isDeviceDebounce(value)">
        <DebounceSettings :id="key" :value="value" />
      </template>
    </div>
  </div>
</template>
