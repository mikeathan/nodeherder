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
    pass in label / value eg disable/0, increase/1, decrease/2
    <div v-for="item in props.items" class="col-xl-2">
        <div class="form-check form-check-inline">
            <input class="form-check-input " type="radio" name="radioSelection" id="radioSelection" :value="item"
                v-model="props.value">
            <label class="form-check-label" for="radioSelection">
                {{ item }}
            </label>
        </div>
    </div>
</template>


