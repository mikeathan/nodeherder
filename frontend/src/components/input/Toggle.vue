<script setup >
import { ref, watch, watchEffect } from "vue";

const emit = defineEmits(['update:value'])

const props = defineProps({
    placeholder:
    {
        type: String,
        default: null
    },
    value: {
        type: Boolean,
        default: null
    }
});

const value = ref(false)


watch(
    () => props.value,
    () => {
        value.value = props.value
    }, { immediate: true }
)
function valueChanged(event) {
    emit('update:value', event);
}

</script>

<template>
    <div>
    </div>
</template>



