<script setup lang="ts">
import { PropType, VNode, computed, h, ref, watch } from "vue";
import { store } from "@/store/index";
import { ExposeTypes } from "@/types/device.type";
import { Expose, ExposeType } from "@/types/device";

const props = defineProps({
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
    disabled: {
        type: Boolean,
        default: false,
        required: false,
    },
});

const emit = defineEmits<{
    (e: "updated", value: any): void;
}>();

const inputValue = ref<any>();

function inputChanged(event: Event) {
    let value: any = (event.target as HTMLInputElement).value;
    if (props.isNumeric) {
        console.log("change convert to number: ", value)
        value = Number(value);
    }

    inputValue.value = value;

    emit('updated', inputValue.value);
}

function isNumber(event: KeyboardEvent) {
    if (props.isNumeric &&
        (!event.key.match(/^[\d\.]$/) ||
            isNaN(Number(inputValue.value)))) {
        console.log("prevent ", event.key)
        event.preventDefault()
    }
}

</script>

<style scoped>
select.form-select,
input.form-control {
    border: 0;
    outline: 0;
    border-radius: 0%;
    border-bottom: 1px solid white;
    text-align: left;
    background-image: none;
}

.form-floating>.form-control~label::after {
    background-color: transparent;
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
</style>
<template>
    <div v-if="props.label != ''" class="form-floating col-sm-3">
        <input type="text" class="form-control" id="dataInput" v-model="inputValue" @input="inputChanged"
            @keypress="isNumber" :disabled="props.disabled" />
        <label for="dataInput">{{ props.label }}</label>
    </div>
    <div v-else class="col-sm-3">
        <input type="text" class="form-control" id="dataInput" v-model="inputValue" @input="inputChanged"
            :disabled="props.disabled" />
    </div>
</template>