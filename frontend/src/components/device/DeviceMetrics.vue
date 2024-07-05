<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref } from 'vue';
import Toggle from '../input/Toggle.vue';
import { DeviceMetricsRequest, DeviceMetrics } from '@/types/metrics';
import { createDeviceSettings } from '@/contracts/settings';
import InputBox from '../input/InputBox.vue';
import { KeyyValuePair } from '@/types/types';

const props = defineProps({
    id: { type: String, required: true },
});
const metricsRequest = ref<KeyyValuePair<DeviceMetricsRequest>>(
    {} as KeyyValuePair<DeviceMetricsRequest>,
);

const deviceMetrics = computed(() => {

    const results = store.getters['metrics/view'](props.id) as DeviceMetrics;
    if (results == null) {
        var request: DeviceMetricsRequest = {
            id: props.id,
            from: 0,
            to: 2
        }

        store.dispatch('metrics/query', request);
    }

    return results;
});


</script>
<template>


</template>
