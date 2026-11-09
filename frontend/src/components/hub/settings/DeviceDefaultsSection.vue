<script setup lang="ts">
  /*
   * Device defaults section (spec 059): supplies the section heading for the shared
   * device settings surface, which is also used per device on the device page.
   */
  import DeviceConfigurationSettings from '@/components/settings/DeviceConfigurationSettings.vue';
</script>

<template>
  <section class="nh-section" aria-labelledby="device-defaults-heading">
    <header class="nh-section-head">
      <h2 id="device-defaults-heading">Device defaults</h2>
      <p class="nh-help">Shared settings for every device that has no override of its own.</p>
    </header>
    <DeviceConfigurationSettings />
  </section>
</template>
