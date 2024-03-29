<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import { DropDownItem } from "../../types/controls.type";

const props = defineProps({
    items: {
        type: Object as PropType<Array<DropDownItem>>,
        default: [],
        required: true
    },
    disabled: {
        type: Boolean,
        default: false,
    }
});

</script>
<!-- 
<button type="button" class="btn" data-bs-toggle="dropdown"
                                aria-expanded="false">
                                <span class="fa fa-plus"></span>

                            </button> -->


<template>
    <button :id="`dropdownControl`" type="button" class="btn" data-bs-toggle="dropdown" aria-expanded="false"
        :disabled="props.disabled">
        <slot></slot>
    </button>
    <ul class="dropdown-menu">
        <li v-for="item in props.items" :key="item.value">
            <a @click="item.event(item.value)" class="dropdown-item" data-toggle="dropdown">
                {{ item.name }}</a>
        </li>
    </ul>
</template>../../types/controls.type