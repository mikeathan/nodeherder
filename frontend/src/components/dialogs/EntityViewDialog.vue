<script setup lang="ts">
  import { ref, watchEffect, computed, onMounted, onUnmounted, watch } from 'vue';
  import LastSeen from '../device/LastSeen.vue';
  import { store } from '../../store/index';
  import { Device, Expose } from '@/types/device';
  import { getFormattedSensorValue, getSensorName } from '../../modules/formatters/sensor-formatter';
  import { ExposeTypes } from '@/types/device.type';
  import { getExposeBinaryProperty, getExposes, toggleExposeBinaryProperty } from '@/contracts/device';
  import {
    writableExposesDeviceFilter,
    writableConfigPresetsExposesDeviceFilter,
  } from '@/configs/automation/device.config';
  import Icon from '../controls/Icon.vue';
  import { EntityInputComponents } from '@/mixins/useEntityComponents';
  import Menu from 'primevue/menu';
  import MenuDropdown from '../controls/MenuDropdown.vue';
  import { getSensorIcon } from '../../modules/formatters/sensor-formatter';
  import { useMiniChartData } from '@/composables/useMiniChartData';
  import { MetricsTypes } from '@/types/metrics.type';
  import MiniNumericChart from '@/components/chart/mini/MiniNumericChart.vue';
  import MiniBinaryChart from '@/components/chart/mini/MiniBinaryChart.vue';

  const props = defineProps<{
    show: boolean;
    title?: string;
    id: string;
    name: string;
  }>();

  const emit = defineEmits(['close']);

  const showDialog = ref<boolean>(props.show);
  const selectedControlExpose = ref<Expose | null>(null);

  // Only fetch chart data when dialog is actually opened
  const { chartData, refetch } = useMiniChartData(props.id, props.name, 24, false);

  watch(
    () => props.show,
    (isShown: boolean) => {
      if (isShown && !chartData.value.hasData && !chartData.value.isLoading) {
        refetch();
      }
    },
    { immediate: true }
  );

  const showMiniChart = computed(() => {
    return (
      chartData.value.hasData && (expose.value.type === ExposeTypes.Numeric || expose.value.type === ExposeTypes.Binary)
    );
  });

  const configExposes = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (!device) {
      return [];
    }

    const exposeNameList = getExposes(device, writableConfigPresetsExposesDeviceFilter());
    return exposeNameList.map((exposeName) => device.exposes[exposeName] as Expose);
  });

  const controlExposes = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (!device) {
      return [];
    }

    const exposeNameList = getExposes(device, writableExposesDeviceFilter());
    return exposeNameList.map((exposeName) => device.exposes[exposeName] as Expose);
  });

  const selectedComponent = computed(() => {
    if (!selectedControlExpose.value) return null;

    return (
      EntityInputComponents(selectedControlExpose.value, 'vertical', {
        update: (e: any) => updateValue(selectedControlExpose.value!.name, e),
      }) ?? null
    );
  });

  watchEffect(() => (showDialog.value = props.show));
  watchEffect(() => {
    // pre select expose control using priority order
    if (!selectedControlExpose.value) {
      const selected = controlExposes.value.reduce<Expose | null>((acc, expose) => {
        if (acc) return acc;

        if (expose.type === ExposeTypes.Numeric) {
          if (expose.values === null) {
            return expose; // Highest priority
          }

          // Lower-priority fallback
          return acc ?? expose;
        }

        return acc;
      }, null);

      selectedControlExpose.value = selected;
    }
  });

  const device = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (!device) return null;

    return device as Device;
  });

  const lastSeen = computed(() => {
    return device.value ? device.value.last_seen : '';
  });

  const isMobile = ref(false);

  const checkMobile = () => {
    if (typeof window !== 'undefined') {
      isMobile.value = window.innerWidth <= 640;
    } else {
      isMobile.value = false;
    }
  };

  const handleEscape = (event: KeyboardEvent) => {
    if (event.key === 'Escape' && showDialog.value) {
      close();
    }
  };

  onMounted(() => {
    checkMobile();
    window.addEventListener('resize', checkMobile);
    window.addEventListener('keydown', handleEscape);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', checkMobile);
    window.removeEventListener('keydown', handleEscape);
  });

  watchEffect(() => (showDialog.value = props.show));

  const expose = computed(() => {
    if (!device.value) return {} as Expose;

    return device.value.exposes[props.name] as Expose;
  });

  function isEnabled() {
    if (device.value?.availability == 'offline') {
      return false;
    }

    if (expose.value.type == ExposeTypes.Numeric && expose.value.data == 0) {
      return false;
    }

    if (controlExposes.value) {
      // temporary fix for state control
      return controlExposes.value.some((expose) => {
        if (expose.type == ExposeTypes.Binary) {
          return getExposeBinaryProperty(expose);
        }
      });
    }
    return true;
  }

  function close() {
    emit('close', false);
    showDialog.value = false;
    selectedControlExpose.value = null;
  }

  const dialogTitle = () => props.title ?? expose.value.name;
  const dialogStyle = computed(() => {
    if (isMobile.value) {
      return {
        width: '100vw',
        height: '100dvh',
        maxHeight: '100dvh',
        margin: '0',
        transform: 'none',
        borderRadius: '0',
        zIndex: '9999',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      };
    }
    return {
      width: 'auto',
      minWidth: '460px',
      maxWidth: '90vw',
      height: 'auto',
      maxHeight: '90vh',
      borderRadius: '1rem',
      overflow: 'auto',
      display: 'flex',
      flexDirection: 'column',
    };
  });

  function handleClick(expose: Expose) {
    if (device.value?.availability == 'offline') {
      return;
    }
    if (expose.type == ExposeTypes.Binary) {
      const value = toggleExposeBinaryProperty(expose);
      updateValue(expose.name, value);
    } else {
      selectedControlExpose.value = expose;
    }
  }

  function updateValue(exposeName: string, newValue: any): void {
    const msg = {
      id: props.id,
      name: exposeName,
      value: newValue,
    };

    store.dispatch('hub/setDeviceValue', msg);
  }
  const menu = ref<InstanceType<typeof Menu> | null>(null);
  const toggleMenu = (event: Event) => {
    menu.value?.toggle(event);
  };

  function buildMenuItems(expose: Expose) {
    if (!expose.values) {
      return [];
    }
    return Object.values(expose.values).map((value: any) => {
      return {
        label: value,
        value: value,
        command: () => {
          updateValue(expose.name, value);
        },
      };
    });
  }
</script>

<template>
  <Dialog
    v-model:visible="showDialog"
    :draggable="false"
    :dismissableMask="true"
    :blockScroll="true"
    :closable="false"
    modal
    :style="dialogStyle"
    @hide="close()">
    <template #header>
      <div class="dialog-header">
        <RouterLink :to="`/devicepage/${id}`">
          <Button label="Link" variant="link" class="ps-0" @click="close()">
            {{ dialogTitle() }}
          </Button>
        </RouterLink>
        <Button icon="pi pi-times" class="p-button-text" @click="close()" />
      </div>
    </template>
    <div class="modal-content-header">
      <div class="modal-value">{{ getFormattedSensorValue(expose) }}</div>
      <LastSeen :timestamp="lastSeen" class="modal-last-seen" />
    </div>

    <!-- Mini chart section -->
    <div v-if="showMiniChart" class="modal-chart-section">
      <MiniNumericChart
        v-if="chartData.type === MetricsTypes.Numeric"
        :data="chartData.data as any"
        :exposeName="expose.name"
        :unit="expose.unit || ''"
        :height="150" />
      <MiniBinaryChart
        v-else-if="chartData.type === MetricsTypes.Binary"
        :data="chartData.data as any"
        :exposeName="expose.name"
        :height="150" />
    </div>

    <div class="modal-content">
      <component
        v-if="selectedComponent"
        :is="selectedComponent"
        :value="selectedControlExpose?.data"
        :disabled="!isEnabled()" />

      <div class="modal-control-buttons">
        <template v-for="expose in controlExposes" :key="expose.name">
          <Icon
            :icon="getSensorIcon(expose.name, expose.data)"
            clickable
            background="#222222"
            :size="38"
            @click="handleClick(expose)" />
        </template>
      </div>
      <div class="modal-config-buttons">
        <template v-for="expose in configExposes" :key="expose.name">
          <MenuDropdown
            backgroundColor="#222222"
            :size="38"
            :text="getSensorName(expose.name)"
            :children="buildMenuItems(expose)"
            @click="toggleMenu"
            :icon="getSensorIcon(expose.name, expose.data)" />
        </template>
      </div>
    </div>
  </Dialog>
</template>
<style scoped>
  .modal-control-buttons {
    display: flex;
    background: #222222;
    border-radius: 999px;
    justify-content: center;
    align-items: center;
    gap: 1rem;
    margin-top: 1.5rem;
    align-self: center;
    width: fit-content;
  }

  .modal-config-buttons {
    display: flex;
    align-items: center;
    gap: 0.2rem;
    padding-top: 0.5rem;
    padding-bottom: 0.5rem;
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    user-select: none;
  }

  .modal-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    user-select: none;
  }

  .modal-content-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.2rem;
    padding-bottom: 1rem;
    user-select: none;
  }

  .modal-chart-section {
    width: 100%;
    margin-bottom: 1.5rem;
  }

  .modal-value {
    font-size: 36px;
    font-weight: 400;
  }
  .modal-last-seen {
    font-size: 16px;
    font-weight: 500;
  }
</style>
