<script setup lang="ts">
  import { ref, computed } from 'vue';
  import { store } from '../../../store/index';
  import { LoggerSettingsType } from '@/types/settings.type';
  import Toggle from '@/components/input/Toggle.vue';
  import Select from 'primevue/select';

  const loggerSettings = computed(() => {
    return store.getters['hub/logger']() as LoggerSettingsType;
  });

  const levelOptions = ref(['trace', 'debug', 'info', 'warn', 'error', 'fatal']);

  function updateSettings(key: keyof LoggerSettingsType, value: any) {
    if (value === loggerSettings.value[key]) {
      return;
    }
    const newSettings = { ...loggerSettings.value, [key]: value };
    store.dispatch('hub/saveLoggerSettings', newSettings);
  }
</script>

<template>
  <h3>Logger</h3>
  <div class="pt-3" />

  <div class="grid grid-nogutter mt-2">
    <dl class="col-12 md:col-3 text-secondary">
      <dt>
        <strong>enableRemoteLogger</strong>
      </dt>
    </dl>
    <div class="col-12 md:col-3">
      <Toggle
        :value="loggerSettings.enableRemoteLogger"
        :valueOn="true"
        :valueOff="false"
        @update="(v: any) => updateSettings('enableRemoteLogger', v)" />
    </div>
  </div>

  <div class="grid grid-nogutter mt-4">
    <dl class="col-12 md:col-3 text-secondary flex align-items-center mb-0">
      <dt>
        <strong>level</strong>
      </dt>
    </dl>
    <div class="col-12 md:col-3">
      <Select
        :modelValue="loggerSettings.level"
        :options="levelOptions"
        @update:modelValue="(v: any) => updateSettings('level', v)"
        class="w-full text-sm" />
    </div>
  </div>
</template>
