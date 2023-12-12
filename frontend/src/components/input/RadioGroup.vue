<script setup>
import { ref, watch } from "vue";

const props = defineProps({
    placeholder: {
        type: String,
        default: ""
    },
    value: null,
    items: Array,

    disabled: Boolean,
});

const emit = defineEmits(['update:data'])

const value = ref(null)

watch(
    () => props.value,
    () => {
        value.value = props.value
    }, { immediate: true }
)

function selectionChanged(event) {

    console.log(event.target.value)
    if (event.target.value == null) {
        return;
    }
    value.value = event.target.value;
    emit("update:data", value.value);
}

</script>

<template>
    <div v-for="item in props.items" class="col-xl-2">
        <div class="form-check form-check-inline">
            <input class="form-check-input " type="radio" name="radioSelection" id="radioSelection" :value="value"
                v-model="props.value" @change="selectionChanged">
            <label class="form-check-label" for="radioSelection">
                {{ item.name }}
            </label>
        </div>
    </div>
</template>


