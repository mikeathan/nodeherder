<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed, ref } from 'vue';
  import Toggle from '../input/Toggle.vue';
  import InputBox from '../input/InputBox.vue';
  import { DeviceConfig } from '@/types/settings.type';
  import { createDeviceConfigOverride } from '@/contracts/settings';
  import { DeviceConfigOverrideComponents } from '@/mixins/useSettingsComponents';
  import { emitOpenConfirmationDialog } from '@/contracts/dialog-events';

  const props = defineProps({
    id: { type: String, required: true },
  });

  const localOverride = ref<DeviceConfig | null>(null);
  const deviceSettings = computed<DeviceConfig | null>({
    get() {
      return localOverride.value ?? store.getters['hub/findDeviceSetting'](props.id) ?? null;
    },
    set(value: DeviceConfig | null) {
      localOverride.value = value;
    },
  });

  const filteredSettings = computed(() => {
    const settings = deviceSettings.value;
    const result: Record<string, any> = {};
    if (!settings) return result;
    
    for (const key in settings) {
      const value = settings[key as keyof DeviceConfig];

      const hasComponent = !!DeviceConfigOverrideComponents[key];
      const isBoolean = typeof value === 'boolean';
      const isPrimitive = typeof value === 'string' || typeof value === 'number';

      if (hasComponent || isBoolean || isPrimitive) {
        result[key] = value;
      }
    }

    return result;
  });

  function createOverride() {
    const newOverride = createDeviceConfigOverride(props.id);
    deviceSettings.value = newOverride;
  }

  function toggleChanged(propName: any, propValue: any) {
    save(propName, propValue);
  }

  function inputLostFocus(propName: any, propValue: any) {
    save(propName, propValue);
  }

  function save(propName: keyof DeviceConfig, propValue: any) {
    if (!deviceSettings.value) return;

    if (!localOverride.value) {
      localOverride.value = { ...deviceSettings.value };
    }

    if (!isObject(localOverride.value[propName])) {
      if (localOverride.value[propName] !== propValue) {
        (localOverride.value as any)[propName] = propValue;
      }
    }

    store.dispatch('hub/saveDeviceConfigOverrides', localOverride.value as DeviceConfig);
  }

  function isObject(value: any): value is object {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
  }

  function deleteOverride() {
    const dlgProps = {
      title: 'Delete',
      message: `Are you sure?`,
    };

    emitOpenConfirmationDialog(() => {
      localOverride.value = null;
      store.dispatch('hub/deleteDeviceConfigOverrides', props.id);
    }, dlgProps);
  }

  function inputUpdated(propName: keyof DeviceConfig, propValue: any) {
    save(propName, { ...propValue });
  }
</script>

<template>
  <div v-if="!deviceSettings">
    <Button @click="createOverride" icon="pi pi-plus" label="Create Override" size="small" />
  </div>
  <div v-else>
    <div class="grid col-12 align-items-center grid-nogutter" v-for="(value, key) in filteredSettings" :key="key">
      <dl class="col-12 md:col-3">
        <dt class="text-secondary">
          <strong> {{ key }}</strong>
        </dt>
      </dl>
      <div class="md:col-4">
        <div v-if="DeviceConfigOverrideComponents[key]">
          <component
            :is="DeviceConfigOverrideComponents[key]"
            v-bind="{
              id: props.id,
              value: value,
            }"
            @update="(v:any) => inputUpdated(key as keyof DeviceConfig, v)" />
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
    <Button label="Delete Override" icon="pi pi-trash" size="small" severity="danger" @click="deleteOverride" />
  </div>
</template>
