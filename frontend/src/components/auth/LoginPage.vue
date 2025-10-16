<script setup lang="ts">
  import { useRouter } from 'vue-router';

  const props = defineProps({
    id: String,
  });
</script>

<template>
  <div class="h-screen flex flex-col justify-center items-center bg-gray-50">
    <img src="/logo.png" alt="MyDashboard Logo" class="w-32 mb-6" />
    <h1 class="text-2xl font-bold mb-4">Welcome to Node-Herder</h1>
    <a href="/auth/google/login" class="flex items-center px-4 py-2 bg-white border rounded shadow hover:bg-gray-100">
      <img src="/google-icon.svg" class="w-6 h-6 mr-2" />
      Sign in with Google
    </a>
    <footer class="mt-10 text-sm text-gray-400">© 2025 Node-Herder</footer>
  </div>
</template>
