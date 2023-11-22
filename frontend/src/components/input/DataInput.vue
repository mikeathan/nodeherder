<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    name: {
        type: String,
        required: true,
    },
    data: Object, // could be string or array of string
    type: String,
    disabled: Boolean
});

const data = ref(null);

const emit = defineEmits(['update:data'])


watch(
    () => props.data,
    () => {
        if (props.type == "binary" && !Array.isArray(props.data)) {
            data.value = []
            console.log("data is the wrong type. Expected type string []")
            return
        }
        data.value = props.data
        // data.value = parseInt(props.data) // ????????? maybe cast it 
    }, { immediate: true }
)

function dataInputChange(event) {
    data.value = event.target.value.replace(/[^0-9.]/g, '');
    emit("update:data", event.target.value);
}

function dataSelectionChanged(event) {
    if (event.target.value == null) {
        return;
    }
    emit("update:data", event.target.value);
}

</script>

<style scoped>
.inputName {
    border: 0;
    outline: 0;
    background: transparent;
    border-bottom: 1px solid #e5e5e5;
    border-radius: 0
}

.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div v-if="type == 'binary'">
        <label class="form-check-label" for="valueinput">Value</label>

        <select id="dataSelect" style="text-align:center;" class="form-control inputName" v-model="data" name="valueinput"
            @change="dataSelectionChanged" :disabled="props.disabled">
            <option v-for="(value, key) in props.data" :value="value" :key="key">
                {{ value }}
            </option>
        </select>
    </div>
    <div v-else>
        <label class="form-check-label" for="valueinput">Value</label>

        <input type="text" class="form-control inputName" name="valueinput" :placeholder="props.name"
            onfocus="this.placeholder = ''" :onblur="this.placeholder = 'props.name'" v-model="data"
            @input="dataInputChange" :disabled="props.disabled">
    </div>
</template>
