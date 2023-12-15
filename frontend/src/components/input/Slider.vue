<script setup lang="ts">
import { ref, watch } from "vue";

const emit = defineEmits<{
    (e: 'change', id: number): void
}>()

const props = defineProps<{
    placeholder?: string
    value?: number,
    min: { type: number, default: 0 }
    max: { type: number, default: 254 }
}>()

function valueChanged(event: Event) {
    var v = parseInt((event.target as HTMLInputElement).value)
    emit('change', v);
}

const value = ref(null)

</script>

<template>
    <div>
        <label for="rangeSelector" class="form-label">Example range</label>
        <input type="range" class="form-range" id="rangeSelector" @change="valueChanged">
    </div>
</template>