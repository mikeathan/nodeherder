<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed, ref } from 'vue';
  import { Device, Expose } from '@/types/device';
  import { ExposeAccessModes, ExposeTypes } from '@/types/device.type';
  import { getExposeAttribute, getExposeProperty, isDeviceOnline } from '../../contracts/device';

  import Toggle from '../input/Toggle.vue';
  import ButtonGroup from '../input/ButtonGroup.vue';
  import { getSensorUnit, getSensorValue } from '@/modules/formatters/sensor-formatter';
  import Range from '../input/Range.vue';
  import Selection from '@/components/input/Selection.vue';

  const props = defineProps({
    id: { type: String, required: true },
  });

  const device = computed(() => {
    return store.getters['hub/findDevice'](props.id) as Device;
  });

  const exposes = computed(() => {
    return device.value ? device.value.exposes : [];
  });

  // TEMPORARY QUICK FIX
  // TODO: do the same we did in Toggle component  so value comes out the correct type eg number
  function updateValue(expose: Expose, value: any) {
    var msg = {
      id: props.id,
      name: expose.name,
      value: value,
    };
    store.dispatch('hub/setDeviceValue', msg);
  }

  // convert keyvaluepair properties to list
  function exposeProperties(expose: Expose): any[] {
    return expose?.values ? Object.values(expose.values) : [];
  }

  // used for expose properties where data seems to be stored as a string or as number
  // but in our selection list we pass int aray of strings
  function convertToString(expose: Expose) {
    if (expose.data != null) return String(expose.data);
    return expose.data;
  }
</script>
<template>
  <div class="grid col-12 align-items-center grid-nogutter" v-for="(expose, index) in exposes" :item="expose">
    <dl class="col-12 md:col-3">
      <dt>
        <strong> {{ expose.name }}</strong>
      </dt>
      <dd class="text-secondary">
        <small> {{ expose.description }} </small>
      </dd>
    </dl>
    <div class="col-12 md:col-9">
      <div v-if="expose.access_mode == ExposeAccessModes.Read">
        {{ getSensorValue(expose.data) }}
        {{ getSensorUnit(expose.name) }}
      </div>

      <div v-else-if="expose.type == ExposeTypes.Numeric" class="align-items-center">
        <!-- TODO: refactor -->
        <div class="pt-3"></div>
        <ButtonGroup
          v-if="expose.values != null"
          :items="(expose.values as any)"
          :value="expose.data"
          @update="(v) => updateValue(expose, v)"
          :disabled="!isDeviceOnline(device)" />
        <div class="pt-3"></div>
        <Range
          :value="expose.data"
          :showInput="true"
          :min="getExposeAttribute(expose, 'min')"
          :max="getExposeAttribute(expose, 'max')"
          @update="(v) => updateValue(expose, v)"
          :disabled="!isDeviceOnline(device)" />
      </div>
      <div v-else-if="expose.type == ExposeTypes.Binary">
        <Toggle
          :value="expose.data"
          :valueOn="getExposeProperty(expose, 'on')"
          :valueOff="getExposeProperty(expose, 'off')"
          @update="(v) => updateValue(expose, v)"
          :disabled="!isDeviceOnline(device)" />
      </div>
      <div v-else-if="expose.type == ExposeTypes.Enum">
        <Selection
          :value="convertToString(expose)"
          :items="exposeProperties(expose)"
          @updated="(v: any) => updateValue(expose, v)"
          :disabled="!isDeviceOnline(device)" />
      </div>
    </div>
  </div>
</template>
