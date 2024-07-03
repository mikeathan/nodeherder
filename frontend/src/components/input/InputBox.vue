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

function isNumber(event: KeyboardEvent) {
    if (props.isNumeric &&
        (!event.key.match(/^[\d\.]$/) ||
            isNaN(Number(inputValue.value)))) {
        event.preventDefault()
    }
}

</script>

<style scoped>
input.form-control {
    border: 0;
    outline: 0;
    border-radius: 0%;
    border-bottom: 1px solid white;
    text-align: left;
    background-color: transparent;
    background-image: none;

}

input.form-control:focus,
:active {
    box-shadow: none;
}


.form-floating>.form-control:focus~label,
.form-floating>.form-control:not(:placeholder-shown)~label,
.form-floating>.form-control~label {
    opacity: 0.6;
    transform: scale(0.85) translateY(-0.7rem) translateX(0.15rem);
}

input.form-control:disabled {
    color: gray;
    background-color: transparent;
}

.form-floating>.form-control~label::after {
    background-color: transparent;
    color: white;

}

.form-floating>.form-control~label {
    color: white;
    background-color: transparent;
}
</style>
<template>
    <div v-if="props.label != ''" class="form-floating">
        <input type="text" class="form-control" id="dataInput" :style="`text-align:${props.position}`"
            v-model="inputValue" @input="inputChanged" @keypress="isNumber" :disabled="props.disabled" />
        <label for="dataInput">{{ props.label }}</label>
    </div>
    <div v-else class="">
        <input type="text" class="form-control" id="dataInput" :style="`text-align:${props.position}`"
            v-model="inputValue" @input="inputChanged" @keypress="isNumber" :disabled="props.disabled" />
    </div>
</template>