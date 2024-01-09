<script setup lang="ts">
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch, PropType } from "vue";
import { getFeatureExposes, getFeatureDevices, createMapFromObject } from "../../modules/convert"
import { ExposeTrigger, Operations, OperationContent } from "../../models/automations"

const store = useStore();

// if expose is numeric
//      operation has steps
// if expose has presets 
//      operations has rotation
// 
// if steps is selected
//      show steps selection
//      set action.operation according to user selection
// if rotation is selected 
//      dont show anything - just set action.operation to 3


const props = defineProps({

    id: {
        type: String,
        required: true
    },
    property: {
        type: String,
        required: true
    },
    operation: {
        type: Number,
        default: 0
    },

});

const id = ref<string>("")
const property = ref<string>("")
const operation = ref<number>(0);
const availableOperations = ref<Array<number>>()
const availableStepOperations = ref<Array<number>>()

const showStepsSelection = ref<boolean>(false)

const emit = defineEmits<{
    (e: 'update:operation', operation: number): void,
}>()

watchEffect(() => id.value = props.id);
watchEffect(() => operation.value = props.operation);

watch(
    () => props.property,
    () => {
        property.value = props.property
        availableOperations.value = tempOperationsBuilder()
    },
    { immediate: true }
);

function tempOperationsBuilder(): Array<number> {
    const availableOperations = Array<number>(0)
    if (feature.value.type === 'numeric') {
        availableOperations.push(1)
    }
    if (feature.value.presets !== undefined) {
        availableOperations.push(3)
    }
    return availableOperations
}

const feature = computed(() => {
    var device = store.getters["devices/find"](id.value);
    if (device == undefined) {
        return []
    }

    if (device.exposes[property.value] == undefined) {
        return []
    }

    return device.exposes[property.value];
});

const getPresets = computed(() => {
    if (feature.value.presets == undefined) {
        return {}
    }

    return feature.value.presets;
})

function operationUpdated(operation: number) {

    switch (operation) {
        case Operations.StepOperation:
            showStepsSelection.value = true;
            return
        case Operations.NoOperation:
        case Operations.RotationOperation:
            showStepsSelection.value = false;
            break;
    }
    emit('update:operation', operation)
}

// testing
function getOperationContext() {
    return Object.values(OperationContent)
}

</script>

<template>
    <div class="col-xl-3 ">
        <Selector :items="availableOperations" :data="operation" @update:data="operationUpdated">
        </Selector>
    </div>
    <div v-if="showStepsSelection" class="col-xl-3 ">
        <Selector :items="getOperationContext" :data="operation" @update:data="operationUpdated">
        </Selector>
    </div>
</template>
