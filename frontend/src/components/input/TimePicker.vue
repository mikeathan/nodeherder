<script setup lang="ts">
import { VTimePicker } from 'vuetify/labs/VTimePicker';
import { ref } from "vue";

const props = defineProps({

    value: {
        type: String,
        default: '',
        required: false,
    },
});

const showTimePicker = ref(false);

const selectedTime = ref<string | null>(props.value);
const emit = defineEmits<{
    (e: 'updated', value: any): void;
}>();

function handleEnterKey() {

    showTimePicker.value = false;
    emit('updated', selectedTime.value);
}

</script>

<template>
    <!-- prepend-icon="mdi-clock-time-four-outline" -->
    <v-text-field dense v-model="selectedTime" label="Start at" readonly @click="showTimePicker = true"
        variant="underlined" />

    <v-dialog v-model="showTimePicker" width="250" @keydown.enter="handleEnterKey">
        <v-time-picker v-model="selectedTime" format="24hr" position="relative" small
            @input="() => (showTimePicker = false)" />
    </v-dialog>

</template>