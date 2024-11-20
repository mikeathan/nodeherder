<script setup lang="ts">
import { ref, watch, PropType } from "vue";
import { SelectSize, SelectFormSize, SelectionItems, LayoutPosition, LayoutPositions } from '@/types/controls.type';

const props = defineProps({
    value: null,
    isNumeric: {
        type: Boolean,
        default: false,
        required: false,
    },
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
    disabled: {
        type: Boolean,
        default: false,
        required: false,
    },
});

const emit = defineEmits<{
    (e: "updated", value: any): void;
    (e: "lostFocus", value: any): void;
}>();


const inputValue = ref<any>(props.value);
watch(
    () => props.value,
    () => {
        inputValue.value = props.value;
    }, { immediate: true }
)

function inputChanged(event: Event) {
    let value: any = (event.target as HTMLInputElement).value;
    if (props.isNumeric) {
        if ((value as string).endsWith('.')) {
            return
        }
        value = Number(value);
    }
    inputValue.value = value;
    emit('updated', inputValue.value);
}

function onLostFocus(event: Event) {
    emit('lostFocus', inputValue.value);
}

function isNumber(event: KeyboardEvent) {
    if (props.isNumeric &&
        (!event.key.match(/^[\d\.]$/) ||
            isNaN(Number(inputValue.value)))) {
        event.preventDefault()
    }
}

</script>


<template>

    <v-text-field dense v-model="inputValue" @input="inputChanged" @keypress="isNumber" :disabled="props.disabled"
        :label="props.label" variant="underlined" />

    <!-- <div v-if="props.label != ''">
             <input type="text" class="form-control" id="dataInput" :style="`text-align:${props.position}`"
            v-model="inputValue" @input="inputChanged" @keypress="isNumber" :disabled="props.disabled"
            :onblur="onLostFocus" />
        <label for="dataInput">{{ props.label }}</label>
    </div>
    <div v-else class="">
        <input type="text" class="form-control" id="dataInput" :style="`text-align:${props.position}`"
            v-model="inputValue" @input="inputChanged" @keypress="isNumber" :disabled="props.disabled"
            :onblur="onLostFocus" />
    </div> -->
</template>