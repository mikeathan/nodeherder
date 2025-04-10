<script setup lang="ts">


import { ref, onMounted, onUnmounted } from 'vue';
import emitter from './event-bus';

// Import all your dialog components
import ConfirmDialog from './dialogs/ConfirmDialog.vue';
import NameDialog from './dialogs/NameDialog.vue';

const currentDialogComponent = ref(null);
const currentDialogProps = ref({});

const dialogMap = {
    confirm: ConfirmDialog,
    name: NameDialog,
    // add more dialog types here
};

function openDialog({ type, props }) {
    currentDialogComponent.value = dialogMap[type];
    currentDialogProps.value = props || {};
}

function closeDialog() {
    currentDialogComponent.value = null;
    currentDialogProps.value = {};
}

onMounted(() => {
    emitter.on('dialog:open', openDialog);
    emitter.on('dialog:close', closeDialog);
});

onUnmounted(() => {
    emitter.off('dialog:open', openDialog);
    emitter.off('dialog:close', closeDialog);
});
</script>


<template>
    <component :is="currentDialogComponent" v-if="currentDialogComponent" v-bind="currentDialogProps"
        @close="closeDialog" />
</template>