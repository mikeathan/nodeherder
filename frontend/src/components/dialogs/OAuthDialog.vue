<script setup lang="ts">
  import { ref, watch, computed } from 'vue';

  const props = defineProps<{
    show: boolean;
    oauthUrl: string;
  }>();

  const emit = defineEmits(['close']);

  const showDialog = ref<boolean>(props.show);

  // Responsive dialog size
  const dialogStyle = computed(() => {
    const isMobile = window.innerWidth <= 768;
    return {
      width: isMobile ? '95vw' : '600px',
      height: isMobile ? '80vh' : '500px',
      maxWidth: '100vw',
      maxHeight: '100vh',
    };
  });

  watch(
    () => props.show,
    (newValue) => {
      showDialog.value = newValue;
    }
  );

  function close() {
    emit('close');
    showDialog.value = false;
  }
</script>

<template>
  <Dialog
    v-model:visible="showDialog"
    modal
    header="Sign in with Google"
    :style="dialogStyle"
    :breakpoints="{ '960px': '75vw', '641px': '95vw' }"
    @hide="close()">
    <div class="oauth-content">
      <p class="oauth-message">Complete your sign-in securely with Google</p>

      <div class="oauth-notice">
        <i class="pi pi-info-circle"></i>
        <span>Google may open additional windows for verification - this is normal</span>
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
  .oauth-content {
    padding: 1rem 0;
  }

  .oauth-message {
    margin: 0 0 1rem 0;
    color: var(--text-color);
    font-size: 1rem;
  }

  .oauth-notice {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem;
    background: var(--surface-100);
    border: 1px solid var(--surface-300);
    border-radius: var(--border-radius);
    color: var(--text-color-secondary);
    font-size: 0.875rem;
  }

  .oauth-notice i {
    color: var(--primary-color);
  }

  /* Mobile specific styles */
  @media (max-width: 768px) {
    .oauth-content {
      padding: 0.5rem 0;
    }

    .oauth-message {
      font-size: 0.9rem;
    }

    .oauth-notice {
      padding: 0.5rem;
      font-size: 0.8rem;
    }
  }
</style>
