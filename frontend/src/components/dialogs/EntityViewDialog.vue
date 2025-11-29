<script setup lang="ts">
  import { ref, watchEffect, computed, onMounted } from 'vue';
  import LastSeen from '../device/LastSeen.vue';
  import { getFormattedSensorValue, getSensorName } from '../../modules/formatters/sensor-formatter';
  import { ExposeTypes } from '@/types/device.type';
  import Icon from '../controls/Icon.vue';
  import Menu from 'primevue/menu';
  import MenuDropdown from '../controls/MenuDropdown.vue';
  import { getSensorIcon } from '../../modules/formatters/sensor-formatter';
  import { useMiniChartData } from '@/composables/useMiniChartData';
  import { useEntityDialog } from '@/composables/useEntityDialog';
  import { useDialogUI } from '@/composables/useDialogUI';
  import { MetricsTypes } from '@/types/metrics.type';
  import MiniChart from '@/components/chart/mini/MiniChart.vue';

  const props = defineProps<{
    show: boolean;
    title?: string;
    id: string;
    name: string;
  }>();

  const emit = defineEmits(['close']);

  const showDialog = ref<boolean>(props.show);

  // Entity dialog logic
  const {
    device,
    expose,
    lastSeen,
    configExposes,
    controlExposes,
    selectedControlExpose,
    selectedComponent,
    isEnabled,
    handleClick,
    buildMenuItems,
  } = useEntityDialog(props.id, props.name);

  // Chart data
  const { chartData, refetch, isLiveUpdates, toggleLiveUpdates } = useMiniChartData(props.id, props.name, 4, false);

  // UI state
  const { dialogStyle } = useDialogUI(() => close());

  const showMiniChart = computed(() => {
    // Don't show chart if there's a brightness control
    const hasBrightnessControl = controlExposes.value.some((exp) => exp.name === 'brightness');
    if (hasBrightnessControl) return false;

    return (
      chartData.value.hasData && (expose.value.type === ExposeTypes.Numeric || expose.value.type === ExposeTypes.Binary)
    );
  });

  // Fetch chart data when dialog is mounted
  onMounted(() => {
    refetch();
  });

  watchEffect(() => (showDialog.value = props.show));

  function close() {
    emit('close', false);
    showDialog.value = false;
    selectedControlExpose.value = null;
  }

  const dialogTitle = () => props.title ?? expose.value.name;

  const menu = ref<InstanceType<typeof Menu> | null>(null);
  const toggleMenu = (event: Event) => {
    menu.value?.toggle(event);
  };
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
      <Button
        :icon="isLiveUpdates ? 'pi pi-sync' : 'pi pi-refresh'"
        :class="['p-button-rounded', 'p-button-text', 'p-button-sm', 'chart-refresh-btn', { active: isLiveUpdates }]"
        @click="toggleLiveUpdates"
        :title="isLiveUpdates ? 'Disable auto refresh' : 'Enable auto refresh'" />
      <MiniChart
        :type="chartData.type"
        :data="chartData.data as any"
        :exposeName="expose.name"
        :unit="expose.unit || ''"
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
    flex-wrap: wrap;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding-top: 0.5rem;
    padding-bottom: 0.5rem;
    max-width: 100%;
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
    position: relative;
  }

  .chart-refresh-btn {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    z-index: 10;
    transition: color 0.2s ease;
  }

  .chart-refresh-btn.active {
    color: #4caf50 !important;
    animation: spin 2s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
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
