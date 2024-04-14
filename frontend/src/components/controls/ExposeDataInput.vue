<script setup lang="ts">
import { computed, ref, watch, PropType } from "vue";
import { store } from "@/store/index";
import { ExposeTypes } from "@/types/device.type";
import { Expose, ExposeType } from "@/types/device";
import InputBox from '@/components/input/InputBox.vue';
import Selection from "@/components/input/Selection.vue";
import { LayoutPosition, LayoutPositions } from '@/types/controls.type';

const props = defineProps({
    id: {
        type: String,
        default: "",
        required: true,
    },
    name: {
        type: String,
        default: '',
        required: true,
    },
    value: null,
    label: {
        type: String,
        default: "",
        required: false,
    },
    position: {
        type: String as PropType<LayoutPosition>,
        default: LayoutPositions.left,
        required: false,
    },
    showPresets: {
        type: Boolean,
        default: false,
        required: false,
    },
    disabled: {
        type: Boolean,
        default: false,
        required: false,
    },
});

const emit = defineEmits<{
    (e: "updated", value: any): void;
}>();

const deviceExpose = computed(() => {
    if (props.name == '') {
        return null;
    }
    var device = store.getters["devices/find"](props.id);
    if (device == undefined) {
        return null;
    }

    return device.exposes[props.name] as Expose;
});

const dataType = computed<ExposeType>(() => deviceExpose.value?.type ?? ExposeTypes.Empty);
const inputValue = ref<any>(props.value);

const selectedPreset = ref<any>('');

const showPresets = computed(() => props.showPresets && exposePresets.value.length != 0);
const exposePresets = computed(() => {
    return deviceExpose.value?.presets == undefined ? [] : deviceExpose.value.presets;
});

watch(
    () => props.name,
    () => {
        inputValue.value = '';
        selectedPreset.value = '';
    }, { immediate: true }
)

function inputChanged(value: any) {
    // reset preset value, if selected
    selectedPreset.value = '';
    inputValue.value = value;
    emit('updated', inputValue.value);
}

const sequenceData = computed(() => {

    if (dataType.value == ExposeTypes.Binary) {
        return ["true", "false"];
    }
    return deviceExpose.value?.attributes ? Object.values(deviceExpose.value.attributes) : [];
});


function sequenceDataSelected(value: any) {
    inputValue.value = value;
    emit('updated', inputValue.value);
}

function presetSelected(selected: any) {
    let value: number = selected;
    try {
        value = parseInt(selected);
    } catch (error) {
        console.error('error converting preset value to number ', error);
    }

    inputValue.value = value;
    emit('updated', inputValue.value);
}

</script>
<template>
    <div v-if="dataType == ExposeTypes.Binary || dataType == ExposeTypes.Enum">
        <Selection :label="props.label" :value="props.value" :disabled="props.disabled" @updated="sequenceDataSelected"
            :position="props.position" :items="sequenceData">
        </Selection>
    </div>
    <div v-if="dataType == ExposeTypes.Numeric">
        <InputBox :label="props.label" :disabled="props.disabled" :value="inputValue" @updated="inputChanged"
            :position="props.position" :isNumeric="true"></InputBox>
    </div>
    <div v-if="showPresets">
        <Selection :label="props.label" :disabled="props.disabled" :value="selectedPreset" @updated="presetSelected"
            :position="props.position" :items="exposePresets">
        </Selection>
    </div>
</template>
