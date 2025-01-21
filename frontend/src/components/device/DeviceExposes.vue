<script setup lang="ts">
import { store } from '../../store/index';
import { computed, ref } from 'vue';
import { Device, Expose } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { getExposeAttribute, getExposeProperty } from '../../contracts/device';

import Toggle from '../input/Toggle.vue';
import ButtonGroup from '../input/ButtonGroup.vue';
import { getSensorUnit, getSensorValue } from '@/modules/formatters/sensor-formatter';
import Range from '../input/Range.vue';

const props = defineProps({
  id: { type: String, required: true },
});

const device = computed(() => {
  return store.getters['hub/findDevice'](props.id) as Device;
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

currently only presets are used in a button grouup
update it so enums are used with button groups too . since we only hanlde features here expect enum properties to list
the data

</script>


<template>
  <div class="grid col-12 align-items-center grid-nogutter" v-for="(expose, index) in device.exposes" :item="expose">
    <dl class="col-12 md:col-3">
      <dt>
        <strong> {{ expose.name }}</strong>
      </dt>
      <dd class="text-secondary">
        <small> {{ expose.description }} </small>
      </dd>
    </dl>
    <div class="col-12 md:col-9">
      <div v-if="expose.properties == null">
        {{ getSensorValue(expose.data) }}
        {{ getSensorUnit(expose.name) }}
      </div>
      <div v-else-if="expose.type == ExposeTypes.Numeric" class="align-items-center">
        <!-- TODO: refactor -->
        <div class="pt-3"></div>

        <ButtonGroup v-if="expose.presets != null" :items="(expose.presets as any)" :value="expose.data"
          @update="(v) => updateValue(expose, v)" />
        <div class="pt-3"></div>
        <Range :value="expose.data" :showInput="true" :min="getExposeAttribute(expose, 'min')"
          :max="getExposeAttribute(expose, 'max')" @update="(v) => updateValue(expose, v)">
        </Range>
      </div>
      <div v-else-if="expose.type == ExposeTypes.Binary">
        <Toggle :minimal="true" :value="expose.data" :valueOn="getExposeProperty(expose, 'on')"
          :value-off="getExposeProperty(expose, 'off')" @update="(v) => updateValue(expose, v)">
        </Toggle>
      </div>
      <div v-else-if="expose.type == ExposeTypes.Enum">WIP : {{ expose.data }}</div>
    </div>
  </div>
</template>
