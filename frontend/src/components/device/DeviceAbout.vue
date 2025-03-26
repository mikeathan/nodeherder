<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed, ref } from 'vue';
  import LastSeen from '../device/LastSeen.vue';
  import PowerSource from '../device/PowerSource.vue';
  import ConnectionType from '../device/ConnectionType.vue';
  import RenameDeviceDialog from '../dialogs/RenameDeviceDialog.vue';
  import ConfirmDialog from '../dialogs/ConfirmDialog.vue';
  import RemoveDeviceDialog from '../dialogs/RemoveDeviceDialog.vue';
  import { Device } from '@/types/device';
  import { RemoveDeviceEvent } from '@/types/dialog.type';
  import { getPowerSourceValue } from '@/contracts/device';

  const props = defineProps({
    id: String,
  });

  const showRenameDialog = ref(false);
  const showInterviewDialog = ref(false);
  const showRemoveDialog = ref(false);

  const device = computed(() => {
    return store.getters['hub/findDevice'](props.id);
  });

  function renameDevice(value: string) {
    store.dispatch('hub/renameDevice', {
      name: device.value.friendly_name,
      newName: value,
    });
  }

  function interviewDevice() {
    store.dispatch('hub/interviewDevice', {
      id: device.value.id,
    });
  }

  function removeDevice(event: RemoveDeviceEvent) {
    console.log('removeDevice', event);
    store.dispatch('hub/removeDevice', {
      id: device.value.id,
      force: event.force ?? false,
      block: event.block ?? false,
    });
  }

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
  <Button icon="pi pi-user-edit" variant="text" v-tooltip="'Rename device'" @click="showRenameDialog = true" />
  <RenameDeviceDialog
    :friendlyName="device.friendly_name"
    :show="showRenameDialog"
    @update:name="renameDevice"
    @close="showRenameDialog = false" />

  <Button icon="pi pi-sync" variant="text" v-tooltip="'Interview device'" @click="showInterviewDialog = true" />
  <ConfirmDialog :show="showInterviewDialog" @confirm="interviewDevice" @close="showInterviewDialog = false" />

  <Button icon="pi pi-trash" variant="text" v-tooltip="'Remove device'" @click="showRemoveDialog = true" />
  <RemoveDeviceDialog
    :friendlyName="device.friendly_name"
    :show="showRemoveDialog"
    @remove="(e) => removeDevice(e)"
    @close="showRemoveDialog = false" />
</template>
