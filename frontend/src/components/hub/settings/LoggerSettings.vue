<script setup lang="ts">
  import { PropType, ref,computed } from 'vue';
  import { store } from '../../../store/index';
  import { LoggerSettingsType, LoggerSettingsTypePropsType } from '@/types/settings.type';
  import Toggle from '@/components/input/Toggle.vue';


  const loggerSettings = computed(() => {
    return store.getters['hub/logger']() as LoggerSettingsType;
  });


  function enableLogging(enabled: any) {
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
        :value="loggerSettings.enableRemoteLogger"
        :valueOn="true"
        :valueOff="false"
        @update="(v: any) => enableLogging(v)" />
    </div>
  </div>
</template>
