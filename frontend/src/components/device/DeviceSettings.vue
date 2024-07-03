<script setup lang="ts">
import { store } from "../../store/index";
import { Ref, computed, onMounted, ref } from "vue";
import Toggle from "../input/Toggle.vue";
import { DeviceSettings } from "@/types/settings";
import { createDeviceSettings } from "@/contracts/settings";
import InputBox from "../input/InputBox.vue";
import { KeyyValuePair } from "@/types/types";

const props = defineProps({
  id: { type: String, required: true },
});
const cachedDeviceSettings = ref<KeyyValuePair<any>>({} as KeyyValuePair<any>);
const isDirty = computed(() => {
  return (
    JSON.stringify(cachedDeviceSettings.value) !==
    JSON.stringify(deviceSettings.value)
  );
});

const deviceSettings = computed(() => {
  if (!store.getters["appconfig/initialized"]() as Boolean) {
    store.dispatch("ws/emit", { event: "loadAppConfig" });
  }

  const settings = store.getters["appconfig/findDeviceSetting"](props.id);

  if (!settings) {
    const newDeviceSettings = JSON.parse(
      JSON.stringify(createDeviceSettings(props.id))
    );
    store.dispatch("appconfig/saveDeviceSettings", newDeviceSettings);
    return newDeviceSettings;
  }
  return JSON.parse(JSON.stringify(settings));
});

onMounted(() => {
  cachedDeviceSettings.value = JSON.parse(JSON.stringify(deviceSettings.value));
  console.log("mounted ", cachedDeviceSettings.value);
});

function updateValue(propName: any, propValue: any) {
  cachedDeviceSettings.value[propName] = propValue;
}

function save() {
  store.dispatch(
    "appconfig/saveDeviceSettings",
    cachedDeviceSettings.value as DeviceSettings
  );
}
</script>
<template>
  <form @submit.prevent="save()">
    <div v-for="(value, key) in deviceSettings" :key="key">
      <label>{{ key }}</label>

      <div v-if="typeof value === 'boolean'">
        <Toggle
          :minimal="false"
          :value="value"
          :valueOn="true"
          :valueoff="false"
          @update="(v) => updateValue(key, v)"
        >
        </Toggle>
      </div>
      <div v-else>
        <InputBox @updated="(v) => updateValue(key, v)" :value="value">
        </InputBox>
      </div>
    </div>
    <button type="button" @click="save()" :disabled="!isDirty">Save</button>
  </form>
</template>
