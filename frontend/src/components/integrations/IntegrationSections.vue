<script setup lang="ts">
  import { computed } from 'vue';
  import { RouterLink } from 'vue-router';
  import {
    integrationSectionLinks,
    type IntegrationNavigationContext,
    type IntegrationSection,
  } from '@/domain/integration-navigation';

  // Presentation owns no instance lookup, provider requests or draft state.
  const props = defineProps<{ active: IntegrationSection; context: IntegrationNavigationContext }>();
  const links = computed(() => integrationSectionLinks(props.context));
</script>

<template>
  <nav class="connection-tabs" aria-label="Integration sections">
    <RouterLink v-for="link in links" :key="link.section" v-slot="{ href, navigate }" :to="link.to" custom>
      <a :href="href" :aria-current="active === link.section ? 'page' : undefined" @click="navigate">{{ link.label }}</a>
    </RouterLink>
  </nav>
</template>
