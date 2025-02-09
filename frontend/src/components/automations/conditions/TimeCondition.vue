<script setup lang="ts">
import { ref, computed, watch, PropType } from 'vue';
import { EqualityOperators } from '../../../contracts/automations';
import { AutomationCondition, TimeCondition } from '../../../types/automation.type.js';
import { store } from '../../../store/index';
import { allExposeFilter } from '@/configs/automation/device.config';
import Selection from '../../input/Selection.vue';
import ExposeSelector from '@/components/controls/ExposeSelector.vue';
import TimePicker from '@/components/input/TimePicker.vue';
import { convertTimeToDate, toHourMinuteString } from '@/contracts/controls';


const props = defineProps({
    item: {
        type: Object as PropType<TimeCondition>,
        default: {} as TimeCondition,
        required: true,
    },
    id: {
        type: String,
        default: '',
        required: true,
    },
});

const condition = ref<TimeCondition>({} as TimeCondition);

const emit = defineEmits<{
    (e: 'update', condition: AutomationCondition): void;
}>();

watch(
    () => props.item,
    () => {
        condition.value = props.item;
    },
    { immediate: true }
);


function exposeSelected(value: string): void {
    if (value == '') {
        return;
    }

    condition.value.name = value;
    condition.value.value = '';
    emit('update', condition.value);
}

function operatorUpdated(operator: string): void {
    condition.value.equality = operator;
    emit('update', condition.value);
}


const timeOperations = computed(() => {
    return EqualityOperators
});

function updateStartAtTime(time: Date): void {
    condition.value.value = toHourMinuteString(time);
}

function hasExposeName(): boolean {
    return condition.value.name != ''
}
</script>

<template>
    <div class="row">
        <div class="col-sm-4">
            <ExposeSelector :id="props.id" :value="condition.name" @updated="exposeSelected" :filter="allExposeFilter()"
                :disabled="hasExposeName()" />
        </div>
        <div class="col-sm-3">
            <Selection :value="condition.equality" @updated="operatorUpdated" :items="timeOperations"
                :disabled="!hasExposeName()" />
        </div>
        <div class="col-sm-5">
            <div class="col-sm-4">
          <TimePicker :value="convertTimeToDate(condition.value)" @updated="(e) => updateStartAtTime(e)"  :disabled="!hasExposeName()"/>
        </div>
        </div>
    </div>
</template>
