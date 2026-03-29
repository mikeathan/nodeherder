<script setup lang="ts">
  import { ref } from 'vue';
  import { isApiResponse } from '@/contracts/api';
  import { useAlerts } from '@/composables/useAlerts';

  const props = defineProps<{
    label: string;
    icon?: string;
    severity?: string;
    action: () => Promise<any>;
  }>();

  const isLoading = ref(false);
  const { showSuccess, showError } = useAlerts();

  function handleResponse(response: any) {
    if (isApiResponse(response) && !response.success) {
      showError(`Action failed: ${response.error}`);
      return;
    }

    showSuccess();
  }

  const handleClick = async () => {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      const result = await props.action();
      handleResponse(result);
    } catch (e) {
      console.error(e);
      showError(`Action failed: ${e}`);
    } finally {
      isLoading.value = false;
    }
  };
</script>

<template>
  <Button
    :label="props.label"
    :icon="props.icon"
    size="small"
    :severity="props.severity || 'primary'"
    :loading="isLoading"
    :disabled="isLoading"
    @click="handleClick" />
</template>
