<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import { DropDownItemType } from "../../types/controls.type";

const props = defineProps({
    items: {
        type: Object as PropType<Array<DropDownItemType>>,
        default: [],
        required: true
    },
    className: {
        type: String,
        default: ""
    },
    disabled: {
        type: Boolean,
        default: false,
    }
});

</script>
<template>
    <button :id="`dropdownControl`" type="button" :class="`btn ${props.className}`" data-bs-toggle="dropdown"
        aria-expanded="false" :disabled="props.disabled">
        <slot></slot>
    </button>
    <ul class="dropdown-menu">
        <li v-for="item in props.items" :key="item.value">
            <a @click="item.event(item.value)" class="dropdown-item" data-toggle="dropdown">
                {{ item.name }}</a>
        </li>
    </ul>
</template>../../types/controls.type