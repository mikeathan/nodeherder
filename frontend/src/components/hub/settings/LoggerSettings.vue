<script setup lang="ts">
  import { PropType, ref } from 'vue';
  import { store } from '../../../store/index';
  import { LoggerSettingsType, LoggerSettingsTypePropsType } from '@/types/settings.type.type';
  import InputBox from '../../input/InputBox.vue';
  import Toggle from '@/components/input/Toggle.vue';

  const props = defineProps({
    settings: {
      type: Object as PropType<LoggerSettingsType>,
      default: {},
      required: true,
    },
  });

  const loggerSettings = ref<LoggerSettingsType>(props.settings);

  // function inputLostFocus(propName: any, propValue: any) {
  //     save(propName, propValue);
  // }

  function save(propName: LoggerSettingsTypePropsType, propValue: any) {
    if (loggerSettings.value[propName] != propValue) {
      loggerSettings.value[propName] = propValue;
      store.dispatch('hub/saveLoggerSettings', loggerSettings.value);
    }
  }

  function enableLogging(enabled: boolean) {
    if (enabled == loggerSettings.value.enableRemoteLogger) {
      return;
    }
    loggerSettings.value.enableRemoteLogger = enabled;
    store.dispatch('hub/saveLoggerSettings', loggerSettings.value);
  }
</script>

<template>
  <h3>Logger</h3>
  <div class="pt-3" />

  <div class="grid grid-nogutter" v-for="(interval, key) in loggerSettings" :key="key">
    <dl class="col-12 md:col-3 text-secondary">
      <dt>
        <strong>{{ key }}</strong>
      </dt>
    </dl>

    <div class="col-12 md:col-3">
      <Toggle
        :minimal="true"
        :value="loggerSettings.enableRemoteLogger"
        :valueOn="true"
        :valueoff="false"
        @update="(v: boolean) => enableLogging(v)" />
    </div>
  </div>
</template>
