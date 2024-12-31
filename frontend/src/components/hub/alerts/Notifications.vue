<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue';
import { store } from '../../../store/index';
import { AlertMessage } from '../../../types/alerts.type';
import Toast from 'primevue/toast';
import { useToast } from 'primevue/usetoast';
import { toToastSeverity } from '@/contracts/alerts';

const toast = useToast();

watchEffect(() => {
  const messages = store.getters[
    'alerts/messages'
  ]() as AlertMessage[];

  messages.forEach((alert: AlertMessage) => {
    if (alert.visible) return;
    alert.visible = true;
    toast.add({
      // @ts-ignore
      severity: toToastSeverity(
        alert.severity!
      ),
      summary: alert.title,
      detail: alert.message,
      life: alert.timeout,
    });
  });
});
</script>
<style scoped></style>

<template>
  <Toast position="top-right" />
</template>
