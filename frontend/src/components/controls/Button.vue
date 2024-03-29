<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import { ButtonType, DropDownItemType } from "../../types/controls.type";

const props = defineProps({

    className: {
        type: String,
        default: ""
    },
    disabled: {
        type: Boolean,
        default: false,
    }
});

const emit = defineEmits<{
    (e: 'click', event: any): void,
}>()

function onClick(event: any): void {

}

</script>
<template>
    <button :class="`btn ${props.className}`" type="button" @click="onClick" :disabled="props.disabled">
        <slot>

        </slot>
    </button>
</template>.