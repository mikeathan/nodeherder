<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue';
import { store } from '../../../store/index';
import { AlertMessage } from '../../../types/alerts.type';
import Toast from 'primevue/toast';
import { useToast } from 'primevue/usetoast';

const toast = useToast();
const messages = computed(() => {
  return store.getters[
    'alerts/messages'
  ]() as AlertMessage[];
});


// const alerts = computed(() => {
//   const alerts = store.getters['alerts/alerts']();

//   alerts.forEach((alert: AlertMessage) => {
//     toast.add({
//       severity: 'success',
//       summary: alert.title,
//       detail: alert.message,
//       life: alert.timeout,
//     });
//   });
// })


function addMessage(alert: AlertMessage) {
  toast.add({
    severity: 'success',
    summary: alert.title,
    detail: alert.message,
    life: alert.timeout,
  });
}

watchEffect(() => {
  // Whenever alerts change, update the component

  const messages = store.getters[
    'alerts/messages'
  ]() as AlertMessage[];
  console.log('watchEffect messages', messages);

  messages.forEach((alert: AlertMessage) => {
    toast.add({
      severity: 'success',
      summary: alert.title,
      detail: alert.message,
      life: alert.timeout,
    });
  })
});

const hasAlerts = computed(() => messages.value.length > 0);

const getSnackbarStyle = (index: number) => {
  return {
    bottom: `${(messages.value.length - index - 1) * 60}px`,
  };
};

let count = 0;
const showError = () => {
  store.dispatch('alerts/showError', 'some test error messsage' + count++);
};

</script>

<template>
  <div class="text-center ma-2">

    <div class="card flex justify-center">
      <Toast />
    </div>

    <Button label="debug show toast" severity="danger" @click="showError()" />
    <!-- <div v-for="(alert, index) in messages" :key="alert.id">
      {{ addMessage(alert) }}

    </div> -->
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
