<script setup lang="ts">
  /*
   * Device settings tab (spec 059): supplies the section heading for the shared device
   * settings surface, matching the Settings → Device defaults section.
   */
  import DeviceConfigurationSettings from '@/components/settings/DeviceConfigurationSettings.vue';
  defineProps<{ id: string }>();
</script>

<template>
  <section class="nh-section" aria-labelledby="device-settings-heading">
    <header class="nh-section-head">
      <h2 id="device-settings-heading">Device settings</h2>
      <p class="nh-help">Use shared defaults, or adjust reporting for this device.</p>
    </header>
    <DeviceConfigurationSettings :id="id" />
  </section>
</template>
