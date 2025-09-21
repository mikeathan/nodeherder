<script setup lang="ts">
  import { ref } from 'vue';

  const props = defineProps<{
    label: string;
    icon?: string;
    action: () => Promise<any>;
  }>();

  const isLoading = ref(false);

  const handleClick = async () => {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      const success = await props.action();
    } catch (e) {
      console.error(e);
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
    severity="primary"
    :loading="isLoading"
    :disabled="isLoading"
    @click="handleClick" />
</template>
