<script setup lang="ts">
  import type { ConnectionFailure } from '@/types/integrations.type';
  function hasField(path: string) {
    return Boolean(document.getElementById(path));
  }
  function focusField(path: string) {
    const element = document.getElementById(path);
    if (!element) return;
    let parent = element.parentElement;
    while (parent) {
      if (parent.tagName === 'DETAILS') (parent as HTMLDetailsElement).open = true;
      parent = parent.parentElement;
    }
    element.tabIndex = -1;
    element.focus();
    element.scrollIntoView({ block: 'center' });
  }
  defineProps<{ error?: ConnectionFailure | null; busy?: boolean }>();
</script>
<template>
  <div v-if="error" class="connection-notice connection-error" role="alert">
    <p>{{ error.message }}</p>
    <p v-if="error.retry_at" class="connection-muted">
      Try again after {{ new Date(error.retry_at).toLocaleString() }}.
    </p>
    <ul v-if="error.field_errors?.length" class="connection-list">
      <li v-for="field in error.field_errors" :key="field.path">
        <button
          v-if="field.path && hasField(field.path)"
          type="button"
          class="nh-btn is-ghost"
          @click="focusField(field.path)">
          {{ field.path }}</button
        ><span v-else-if="field.path">{{ field.path }}:</span> {{ field.message }}
      </li>
    </ul>
  </div>
  <p v-if="busy" role="status" class="connection-muted">Working…</p>
</template>
