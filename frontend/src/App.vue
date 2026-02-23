<script setup lang="ts">
  import AuthWrapper from '@/components/layout/AuthWrapper.vue';
  import { useTheme } from '@/services/theme.service';
  import { onMounted } from 'vue';

  const { initTheme } = useTheme();

  onMounted(() => {
    initTheme();
  });
</script>

<template>
  <AuthWrapper />
</template>
<style>
  html,
  body {
    margin: 0;
    padding: 0;
    height: 100%;
    overflow: hidden;
  }

  #app {
    height: 100vh;
    display: flex;
    flex-direction: column;
  }

  /* Mobile: enable native body scroll so browser chrome (URL bar) can hide */
  @media screen and (max-width: 768px) {
    html,
    body {
      height: auto !important;
      min-height: 100dvh !important;
      overflow: visible !important;
    }
    #app {
      height: auto !important;
      min-height: 100dvh !important;
    }
  }

  body {
    font-family: 'Roboto', sans-serif !important;
  }

  .p-component {
    font-family: 'Roboto', sans-serif !important;
  }
</style>
