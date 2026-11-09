<script setup lang="ts">
  /*
   * Appearance section (spec 059): supplies the section heading for the shared
   * appearance customiser, which the header drawer also renders without one.
   */
  import AppearanceForm from '@/components/shell/AppearanceForm.vue';
</script>

<template>
  <section class="nh-section" aria-labelledby="appearance-heading">
    <header class="nh-section-head">
      <h2 id="appearance-heading">Appearance</h2>
      <p class="nh-help">Theme, accent, density and text size are saved in this browser only.</p>
    </header>
    <AppearanceForm />
  </section>
</template>
