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

function change(event) {

}
</script>

<template>
    <div v-for="item in props.items" class="col-xl-2">
        <div class="form-check form-check-inline">
            <input class="form-check-input " type="radio" name="radioSelection" id="radioSelection" :value="item.value"
                v-model="props.value" @change="change">
            <label class="form-check-label" for="radioSelection">
                {{ item.name }}
            </label>
        </div>
    </div>
</template>


