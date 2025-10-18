<script setup lang="ts">
import { useAuth } from '@/mixins/composables/useAuthentication';
import LoginPage from '@/components/auth/LoginPage.vue';
import MainLayout from '@/components/layout/MainLayout.vue';

const { isAuthenticated } = useAuth();
</script>

<template>
  <LoginPage v-if="!isAuthenticated" />
  <MainLayout v-else />
</template>
