<script setup lang="ts">
import { store } from '../../store/index';
import { computed, onMounted, ref } from 'vue';
import {
    DeviceMetricsRequest,
    DeviceMetrics,
} from '@/types/metrics';
import Selection from '../input/Selection.vue';
import { KeyyValuePair } from '@/types/types';
import { toUnix } from '@/utils/date.utils';

const props = defineProps({
    id: { type: String, required: true },
});

const historySelection: string[] = [
    'last hour',
    'last day',
    'last week',
    'last month',
];
const timeOffsets: KeyyValuePair<number> = {
    'last hour': 1,
    'last day': 24,
    'last week': 168,
    'last month': 720,
};
const fromDate = ref<Date | null>(null);
const toDate = ref<Date | null>(null);

function dateSelected(value: any) {
    const dateOffset = timeOffsets[value];

    toDate.value = new Date();
    fromDate.value = new Date();
    fromDate.value.setHours(
        toDate.value.getHours() - dateOffset,
    );

    var request: DeviceMetricsRequest = {
        id: props.id,
        from: toUnix(fromDate.value ?? new Date()), // temp
        to: toUnix(toDate.value ?? new Date()), // temp
    };

    store.dispatch('metrics/query', request);
}

const metricsRequest = ref<
    KeyyValuePair<DeviceMetricsRequest>
>({} as KeyyValuePair<DeviceMetricsRequest>);


const deviceMetrics = computed(() => {
    const results = store.getters['metrics/view'](
        props.id,
    );
    // if (results == null) {
    //     var request: DeviceMetricsRequest = {
    //         id: props.id,
    //         from: toUnix(fromDate.value ?? new Date()), // temp
    //         to: toUnix(toDate.value ?? new Date()), // temp
    //     };

    //     store.dispatch('metrics/query', request);
    // }

    return results;
});


</script>
<style scoped></style>

<template>
    <div class="col-sm-3">
        <Selection label="Select time offset:" @updated="dateSelected" :items="historySelection">
        </Selection>
    </div>
    {{ deviceMetrics }}
</template>
