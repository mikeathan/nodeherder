<script setup lang="ts">

type IconType = {
    name: string;
    color: string;
}

const props = defineProps({

    icon: {
        type: Object as PropType<IconType>,
        default: {} as IconType,
    },
    name: {
        type: String,
        default: '',
        required: true,
    },
    color: {
        type: String,
        default: '#4b5563',
        required: false,
    }
});
</script>
<style scoped>
.icon {
    width: 24px;
    height: 24px;
    vertical-align: middle;
    margin-right: 0.5em;
}
</style>
<template>
    <svg viewBox="0 0 24 24" class="icon  " :style="{ fill: props.color }">
        <path :d="props.name" />
    </svg>
</template>
