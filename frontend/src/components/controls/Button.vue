<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import { ButtonType, DropDownItemType } from "../../types/controls.type";

const props = defineProps({
    item: {
        type: Object as PropType<ButtonType>,
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
    <button class="btn btn-light " type="button" @click="saveAction" :disabled="props.disabled">
        <slot>

        </slot>
    </button>
</template>.