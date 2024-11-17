<script setup lang="ts">
import { computed } from 'vue';
import { key, store } from '../../../store/index';
import { AlertMessage } from '../../../types/alerts.type';


const messages = computed(() => {
  return store.getters[
    'alerts/messages'
  ]() as AlertMessage[];
});

</script>

<template>
  <div class="text-center ma-2">

    NOTIFICATIONS TEST
    <v-snackbar v-for="(message, index) in messages" :key="message.id" :timeout="message.timeout" :top="true"
      @input="() => store.dispatch('alerts/removeAlert', message.id)">
      {{ message.message }}


      <template v-slot:actions>
        <v-btn color="white" @click="store.dispatch('alerts/removeAlert', index)">
          Dismiss
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<!-- https://vuetifyjs.com/en/components/snackbars/#usage -->