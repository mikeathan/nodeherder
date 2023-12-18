<script setup >
import { ref, watch } from "vue";

const emit = defineEmits(['update:value'])

const props = defineProps({
    placeholder:
    {
        type: String,
        default: null
    },
    enabled: {
        type: Boolean,
        default: null
    }
});

const enabled = ref(false)

watch(
    () => props.enabled,
    () => {
        enabled.value = props.value
    }, { immediate: true }
)

function valueChanged(event) {
    emit('update:value', event);
}

</script>
<style scoped>
.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div>
        <label v-if="props.placeholder != null" class="form-check-label">{{ props.placeholder }}</label>
        <input class="form-check-input custom-control-input" type="checkbox" role="switch" id="flexSwitchCheckDefault"
            v-model="enabled" @change="valueChanged">
    </div>
</template>



