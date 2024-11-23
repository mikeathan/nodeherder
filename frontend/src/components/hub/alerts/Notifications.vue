<script setup lang="ts">
import { computed } from 'vue';
import { store } from '../../../store/index';
import { AlertMessage } from '../../../types/alerts.type';

const messages = computed(() => {
  return store.getters[
    'alerts/messages'
  ]() as AlertMessage[];
});

const hasAlerts = computed(() => messages.value.length > 0);

const getSnackbarStyle = (index: number) => {
  return {
    bottom: `${(messages.value.length - index - 1) * 60}px`,
  };
};
</script>

<template>
  <div class="text-center ma-2">
    <!-- <v-snackbar v-for="(alert, index) in messages" :key="alert.id" :color="alert.color" :timeout="alert.timeout"
      location="top" :style="getSnackbarStyle(index)" v-model="hasAlerts"
      @input="() => store.commit('alerts/removeAlert', alert.id)">
      {{ alert.message }}

      <template v-slot:actions>
        <v-btn color="white" @click="store.commit('alerts/removeAlert', alert.id)">
          Dismiss
        </v-btn>
      </template>
    </v-snackbar> -->
  </div>
</template>
