<script setup lang="ts">
  /* Errors for one editor field, matched by exact path (domain/automation.ts Issue paths). */
  import { computed } from 'vue';
  import { Issue } from '@/domain/automation';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{ issues: Issue[]; path: string }>();
  const own = computed(() => props.issues.filter((i) => i.path === props.path));
</script>

<template>
  <p v-for="i in own" :key="i.message" class="nh-err"><UiIcon name="error" />{{ i.message }}</p>
</template>
