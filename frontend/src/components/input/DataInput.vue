<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    placeholder: {
        type: String,
        default: ""
    },
    data: null, // could be string or array of string
    items: null,
    type: {
        type: String,
        default: "numeric"
    },
    disabled: Boolean,
    width: {
        type: Number,
        default: 0
    },
    alignment: {
        type: String,
        default: 'center'
    }
});

const data = ref(null);
const placeholder = ref("");
const items = ref([])
const emit = defineEmits(['update:data'])


watch(
    () => props.data,
    () => {
        data.value = props.data
        if ((props.type == 'numeric') && isNaN(data.value)) {
            data.value = ""
        }
    }, { immediate: true }
)

watch(
    () => props.items,
    () => {
        items.value = props.items
        if (items.value == null) {
            items.value = []
        }

    }, { immediate: true }
)

watch(
    () => props.placeholder,
    () => {
        placeholder.value = props.placeholder
    }, { immediate: true }
)

function isSelection() {
    return items.value.length > 0
}

function dataInputChange(event) {
    if (props.type == 'numeric') {
        var value = event.target.value.replace(/[^0-9.]/g, '');
        data.value = parseInt(value)
    }
    emit("update:data", data.value);
}

function dataSelectionChanged(event) {
    if (event.target.value == null) {
        return;
    }
    if (typeof data.value === "boolean") {
        data.value = Boolean(event.target.value).valueOf()
    } else {
        data.value = event.target.value
    }

    emit("update:data", data.value);
}

function blurChanged(event) {
    placeholder.value = props.placeholder
}

function focusChanged(event) {
    placeholder.value = ""
}

</script>

<style scoped>
select.form-control,
input.form-control {
    border: 0;
    outline: 0;
    border-radius: 0%;
    border-bottom: 1px solid white;
}


select.form-control:disabled,
input.form-control:disabled {
    border: 0;
    outline: 0;
    border-radius: 0%;
    background-color: transparent;
    border-bottom: 0px solid white;
}


input.form-control:focus,
:active,
:hover,
select.form-control:focus,
:active,
:hover {
    box-shadow: none;
}

.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div v-if="isSelection()">
        <select id="dataSelect" :style="'text-align:' + props.alignment + ';'" class="form-control " name="valueinput"
            v-model="data" @change="dataSelectionChanged" :disabled="props.disabled">
            <option v-if="props.placeholder != ''" value="">{{ props.placeholder }}</option>
            <option v-for="(value, key) in items" :value="value" :key="key">
                {{ value }}
            </option>
        </select>
    </div>
    <div v-else>
        <input type="text" class="form-control" :style="'text-align:' + props.alignment + ';'" name="valueinput"
            :placeholder="placeholder" v-model="data" @input="dataInputChange" :disabled="props.disabled"
            @focus="focusChanged" @blur="blurChanged">
    </div>
</template>
