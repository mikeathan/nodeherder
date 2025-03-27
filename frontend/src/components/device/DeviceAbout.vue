<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed } from 'vue';
  import LastSeen from '../device/LastSeen.vue';
  import PowerSource from '../device/PowerSource.vue';
  import ConnectionType from '../device/ConnectionType.vue';
  import { Device } from '@/types/device';
  import { getPowerSourceValue } from '@/contracts/device';
  import DeviceControl from './DeviceControl.vue';

  const props = defineProps({
    id: String,
  });

  const displayProps = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (device == undefined) {
      return [];
    }

    return [
      {
        key: 'IEEE Address:',
        value: device.id,
      },
      {
        key: 'Friendly name:',
        value: device.friendly_name,
      },
      {
        key: 'Description:',
        value: device.description,
      },
      {
        key: 'Availability:',
        value: device.availability,
      },
      {
        key: 'Last seen:',
        type: LastSeen,
        props: {
          timestamp: device.last_seen,
        },
      },
      {
        key: 'Power source:',
        type: PowerSource,
        props: {
          power_source: device.power_source,
          value: getPowerSourceValue(device),
        },
      },
      {
        key: 'Connection Type:',
        type: ConnectionType,
        props: {
          type: device.connection_type,
        },
      },
    ];
  });
</script>
<template>
  <div>
    <dl class="grid grid-nogutter" v-for="(prop, idx) in displayProps" :key="idx">
      <dt class="col-12 md:col-5 text-secondary">
        {{ prop.key }}
      </dt>
      <dd class="col-12 md:col-7">
        <div v-if="prop.type === undefined">
          <span title="last update">{{ prop.value }}</span>
        </div>
        <template v-else>
          <component :is="prop.type" v-bind="prop.props"></component>
        </template>
      </dd>
    </dl>
  </div>
  <DeviceControl :id="props.id" />
</template>
