<script setup>
import { ref, watch } from "vue";

const props = defineProps({

    value: null,
    items: Array,
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

    if (event.target.value == null) {
        return;
    }
    value.value = event.target.value;
    emit("update:data", value.value);
}

</script>

<template>
    <div class="col-50">
        <div v-for="item in props.items" class="btn-group">
            <input type="radio" class="btn-check" name="options-outlined" :id="`radioSelection${item.name}`"
                :value="item.value" :checked="props.value == item.value" @change="selectionChanged">
            <label class="btn btn-outline-secondary" :for="`radioSelection${item.name}`"> {{ item.name }}</label>
        </div>
    </div>
</template>


