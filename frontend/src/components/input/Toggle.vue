<script setup lang="ts">
import { ref, watch } from "vue";

const emit = defineEmits<{
    (e: 'update:value', value: Boolean): void,
}>()

const props = defineProps({
    placeholder:
    {
        type: String,
        default: null
    },
    enabled: {
        type: Boolean,
        default: false
    },
    scale: {
        type: Number,
        default: 1.0
    },
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
<style scoped>
.custom-control-input {
    transform: scale(1.1);
}
</style>
<template>
    <div class=" form-check form-switch">
        <label v-if="props.placeholder != null" class="form-check-label">{{ props.placeholder }}</label>
        <input class="form-check-input custom-control-input" type="checkbox" role="switch" id="flexSwitchCheckDefault"
            v-model="enabled" @change="valueChanged" :disabled="enabled == null">
    </div>
</template>



