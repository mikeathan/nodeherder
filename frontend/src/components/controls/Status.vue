<script setup lang="ts">
  import { KeyValuePair } from '@/types/types';
  import { store } from '../../store/index';
  import { computed, watch, ref } from 'vue';
  import { ConnectionStateIcon } from '@/types/connection.type';

  const connectionStatus = computed(() => {
    return store.getters['ws/getConnectionStatus'];
  });

  const connectionStatusState: KeyValuePair<ConnectionStateIcon> =
    {
      connected: {
        icon: 'pi pi-circle-fill',
        color: 'green',
      },
      disconnected: {
        icon: 'pi pi-circle-fill',
        color: 'red',
      },
      connecting: {
        icon: 'pi pi-spin pi-spinner',
        color: 'white',
      },
    };

  const status = ref<ConnectionStateIcon>(
    connectionStatusState['disconnected']
  );

  watch(
    () => connectionStatus.value,
    () => {
      status.value =
        connectionStatusState[connectionStatus.value];
    },
    { immediate: true }
  );
</script>

<template>
  <i
    :class="status.icon"
    :style="{
      color: status.color,
      fontSize: '0.5rem',
    }"></i>
</template>
