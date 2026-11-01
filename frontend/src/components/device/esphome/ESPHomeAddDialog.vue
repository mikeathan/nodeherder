<script setup lang="ts">
  /*
   * ESPHome onboarding dialog. Form state, discovery and submission live in
   * useESPHomeAddDialog.ts; this component renders the form.
   */
  import { ESPHOME_MAX_NAME_LENGTH, ESPHOME_PORT_MAX, ESPHOME_PORT_MIN } from '@/contracts/esphome';
  import { ESPHOME_DEFAULT_DEVICE_LABEL } from '@/domain/esphome';
  import { useESPHomeAddDialog } from '@/composables/esphome/useESPHomeAddDialog';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  defineProps<{ open: boolean }>();
  const emit = defineEmits<{ (event: 'close'): void; (event: 'added'): void }>();
  const {
    name,
    address,
    port,
    key,
    busy,
    scanning,
    scanned,
    candidates,
    error,
    close,
    discover,
    selectCandidate,
    submit,
  } = useESPHomeAddDialog({
    onClose: () => emit('close'),
    onAdded: () => emit('added'),
  });
</script>

<template>
  <UiDialog :open="open" title="Add ESPHome device" :dismissible="!busy" @close="close">
    <p>Enter the device's private IP address and Native API port. Zigbee devices continue to use Permit join.</p>
    <UiButton icon="search" :loading="scanning" :disabled="busy" @click="discover">Discover ESPHome devices</UiButton>
    <p v-if="scanned && !candidates.length" class="nh-muted">No candidates found. You can enter an address manually.</p>
    <ul v-if="candidates.length">
      <li v-for="candidate in candidates" :key="`${candidate.address}:${candidate.port}`">
        <UiButton size="sm" @click="selectCandidate(candidate)">{{ candidate.name || ESPHOME_DEFAULT_DEVICE_LABEL }} · {{ candidate.address }}:{{ candidate.port }}</UiButton>
      </li>
    </ul>
    <form id="esphome-add-form" class="nh-form" @submit.prevent="submit">
      <label class="nh-field" for="esphome-name">Name <input id="esphome-name" v-model="name" class="nh-input" :maxlength="ESPHOME_MAX_NAME_LENGTH" autocomplete="off" /></label>
      <label class="nh-field" for="esphome-address">IP address <input id="esphome-address" v-model="address" class="nh-input" required autocomplete="off" placeholder="192.168.1.10" /></label>
      <label class="nh-field" for="esphome-port">Native API port <input id="esphome-port" v-model.number="port" class="nh-input" type="number" :min="ESPHOME_PORT_MIN" :max="ESPHOME_PORT_MAX" required /></label>
      <label class="nh-field" for="esphome-key">Encryption key (optional) <input id="esphome-key" v-model="key" class="nh-input" type="password" autocomplete="off" spellcheck="false" /></label>
      <p v-if="error" role="alert">{{ error }}</p>
    </form>
    <template #footer>
      <UiButton :disabled="busy" @click="close">Cancel</UiButton>
      <UiButton type="submit" variant="primary" :loading="busy" :disabled="scanning" form="esphome-add-form">Add device</UiButton>
    </template>
  </UiDialog>
</template>
