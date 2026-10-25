<script setup lang="ts">
  import { focusConnectionField } from './field-focus';
  import type { ConnectionFailure } from '@/types/integrations.type';
  function hasField(path: string) {
    return Boolean(document.getElementById(path));
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
          @click="focusConnectionField(field.path)">
          {{ field.path }}</button
        ><span v-else-if="field.path">{{ field.path }}:</span> {{ field.message }}
      </li>
    </ul>
  </div>
  <p v-if="busy" role="status" class="connection-muted">Working…</p>
</template>
