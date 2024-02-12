<script setup lang="ts">
import { ref, watch } from "vue";

const emit = defineEmits<{
    (e: 'update:value', value: Boolean): void,
}>()

const props = defineProps({
    showLabels:
    {
        type: Boolean,
        default: false
    },
    enabled: {
        type: Boolean,
        default: false
    }
});

const enabled = ref(false)

watch(
    () => props.enabled,
    () => {
        enabled.value = props.enabled
    }, { immediate: true }
)

function valueChanged(): void {
    emit('update:value', enabled.value);
}

</script>
<template>
    <button v-if="showLabels" type="button" class="btn btn-link">OFF</button>
    <div class="form-check form-switch form-check-inline align-middle me-0">
        <input class="form-check-input" type="checkbox" v-model="enabled" @change="valueChanged"
            :disabled="enabled == null">
    </div>
    <button v-if="showLabels" type="button" class="btn btn-link">ON</button>
</template>



