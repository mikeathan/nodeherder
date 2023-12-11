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



</script>

<template>
    <div v-for="item in props.items">
        <div class="form-check form-check-inline">
            <input class="form-check-input " type="radio" :name="`radio${item}`" :id="`radio${item}`" :value="item"
                v-model="props.value">
            <label class="form-check-label" :for="`radio${item}`">
                {{ item }}
            </label>
        </div>
    </div>
</template>


