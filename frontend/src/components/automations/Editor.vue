<script setup lang="ts">
  import { useRouter } from 'vue-router';
  import DeviceAutomation from './DeviceAutomation.vue';

  const props = defineProps({
    id: String,
  });

  const router = useRouter();
  function cancel(): void {
    router.push('/viewer');
  }
</script>

<template>
  <Card>
    <template #title>
      <h2>Automation Editor</h2>
    </template>
    <template #content>
      <DeviceAutomation :id="props.id" @cancel="cancel" />
    </template>
  </Card>
</template>
