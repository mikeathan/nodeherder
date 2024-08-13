<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref, watch } from 'vue';
import Toggle from '../input/Toggle.vue';
import { DeviceSettings } from '@/types/settings';
import { createDeviceSettings } from '@/contracts/settings';
import InputBox from '../input/InputBox.vue';
import { KeyValuePair } from '@/types/types';

const props = defineProps({
  id: { type: String, required: true },
});

const deviceSettings = computed(() => {
  if (!store.getters['appconfig/initialized']() as Boolean) {
    store.dispatch('ws/emit', { event: 'loadAppConfig' });
  }

  const settings = store.getters['appconfig/findDeviceSetting'](props.id);
  return settings
});


TODO - save on form change

function updateValue(propName: any, propValue: any) {
  deviceSettings.value[propName] = propValue;
  console.log("updateValue", deviceSettings.value);
}

function save() {
  console.log("save", deviceSettings.value);
  store.dispatch('appconfig/saveDeviceSettings', deviceSettings.value as DeviceSettings);
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
        <Toggle :minimal="false" :value="value" :valueOn="true" :valueoff="false" @update="(v) => updateValue(key, v)">
        </Toggle>
      </div>
      <div v-else>
        <InputBox @updated="(v) => updateValue(key, v)" :value="value" :disabled="typeof value !== 'number'"
          :is-numeric="typeof value === 'number'">
        </InputBox>
      </div>
    </div>
  </div>
  <br />
  <div class="pb-3">
    <button type="button" class="btn btn-light" aria-label="Save" @click="save()">
      Save
    </button>
  </div>
</template>
