<script setup lang="ts">
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch, PropType } from "vue";
import { getFeatureExposes, getFeatureDevices, createMapFromObject } from "../../modules/convert"
import { Operations, ExposeTrigger } from "../../models/automations"

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
});

const id = ref<string>()
const property = ref<string>()

const availableOperations = ref<Array<number>>()


const emit = defineEmits<{
    (e: 'update:operation', operation: number): void,
}>()

watchEffect(() => id.value = props.id);
watchEffect(() => property.value = props.property);

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


</script>

<template>
    <div class="col-xl-3 ">
        <Selector :items="availableOperations">
        </Selector>
    </div>
</template>
