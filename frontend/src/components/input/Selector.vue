<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    placeholder: {
        type: String,
        default: ""
    },
    value: null, // could be string or array of string
    items: Object,
    disabled: Boolean,
    alignment: {
        type: String,
        default: 'center'
    }
});

const selected = ref(null);
const emit = defineEmits(['update:data'])

watch(
    () => props.value,
    () => {
        selected.value = props.value

    }, { immediate: true }
)

function dataSelectionChanged(event) {
    if (event.target.value == null) {
        return;
    }

    if (typeof selected.value === "boolean") {
        selected.value = Boolean(event.target.value).valueOf()
    } else {
        selected.value = event.target.value
    }

    emit("update:data", selected.value);
}

</script>

<style scoped>
maybe use bottom lines like before select.form-control {
    appearance: none;
    border: none;
    outline: none;
    background: none;
    background-color: transparent;

}

select.form-control:disabled {
    border: 1;
    outline: 1;
    outline-color: white;
    background: #545968;

}

select.form-control:focus,
:active,
:hover {
    outline: none;
    box-shadow: none;
}

/* 

select.form-control {
    border: 1;
    outline: 1;
    appearance: none;
    border: none;
    background: none;
    background-color: transparent;
    font-family: inherit;
    outline: none;
}

.plain-select:focus,
:active,
:hover {
    outline: none;
    box-shadow: none;
}


select.form-control:focus,
:active,
:hover {
    outline: none;
    box-shadow: none;
}

select.form-control {
    background-color: transparent;
    background-clip: padding-box;
    border: 1;
    outline: 1;
    appearance: none;
}

select.form-control {
    background-color: transparent;
    background-clip: padding-box;
} */
</style>
<template>
    <div>
        <select id="dataSelect" :style="'text-align:' + props.alignment + ';'" class="form-control" v-model="selected"
            @change="dataSelectionChanged" :disabled="props.disabled">
            <option v-if="props.placeholder != ''" value="">{{ props.placeholder }}</option>
            <option v-for="(value, key) in props.items" :value="value" :key="value">
                {{ key }}
            </option>
        </select>
    </div>
</template>
