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
        severity: toToastSeverity(
          alert.severity!
        ) as typeof ToastMessageOptions, // TODO - fix
        summary: alert.title,
        detail: alert.message,
        life: alert.timeout,
      });
    });
  });
</script>

<template>
  <div class="text-center ma-2">
    <Toast />
  </div>
</template>
