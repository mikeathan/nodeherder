<script setup lang="ts">
  /*
   * ESPHome connection management dialog. Polling, edit drafts and fenced
   * mutations live in useESPHomeManageDialog.ts; this component renders state.
   */
  import { ESPHOME_MAX_NAME_LENGTH, ESPHOME_PORT_MAX, ESPHOME_PORT_MIN } from '@/contracts/esphome';
  import { ESPHOME_UNKNOWN_STATUS } from '@/domain/esphome';
  import { useESPHomeManageDialog } from '@/composables/esphome/useESPHomeManageDialog';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ open: boolean }>();
  const emit = defineEmits<{ (event: 'close'): void }>();
  const {
    items,
    loading,
    loaded,
    loadError,
    error,
    removing,
    editing,
    editName,
    editAddress,
    editPort,
    keyMode,
    editKey,
    saving,
    changing,
    unmatchedDevices,
    displayName,
    remove,
    edit,
    closeEdit,
    save,
    diagnosticMessage,
    initialized,
  } = useESPHomeManageDialog(() => props.open);
</script>

<template>
  <UiDialog :open="open" title="Manage ESPHome devices" @close="emit('close')">
    <p class="nh-muted">Connection settings and status for your ESPHome devices. Connected devices appear in the device list automatically. Encryption keys are never shown.</p>
    <p v-if="loadError" role="alert">{{ loadError }}<template v-if="loaded"> Showing the last loaded settings.</template></p>
    <p v-if="error" role="alert">{{ error }}</p>
    <p v-if="loading && !loaded" aria-busy="true">Loading devices…</p>
    <p v-else-if="loaded && !initialized" aria-busy="true">Checking registered devices…</p>
    <p v-else-if="loaded && initialized && !items.length && !unmatchedDevices.length" class="nh-muted">No ESPHome devices have been added.</p>
    <div v-if="loaded && initialized && unmatchedDevices.length" role="status">
      <p>Some registered ESPHome devices have no matching saved connection settings. Their device records are retained, but connection settings cannot be managed here.</p>
      <ul><li v-for="device in unmatchedDevices" :key="device.id">{{ device.friendly_name }}</li></ul>
      <p class="nh-muted">Check the device-provider data and keyring configuration before adding the device again.</p>
    </div>
    <div v-for="item in items" :key="item.id" class="nh-provider-row">
      <div>
        <strong>{{ displayName(item) }}</strong><br />
        <span class="nh-muted">{{ item.address }}:{{ item.port }} · {{ item.status || ESPHOME_UNKNOWN_STATUS }}<template v-if="item.reconnections"> · {{ item.reconnections }} reconnection attempts</template></span>
        <p v-if="item.diagnostic" class="nh-muted">{{ diagnosticMessage(item.diagnostic) }}</p>
      </div>
      <div>
        <UiButton :disabled="changing" variant="ghost" icon="edit" :label="`Edit ${displayName(item)}`" @click="edit(item)" />
        <UiButton :disabled="changing" variant="ghost" icon="delete" :label="`Remove ${displayName(item)}`" :loading="removing === item.id" @click="remove(item)" />
      </div>
    </div>
    <form v-if="editing" class="nh-form" @submit.prevent="save">
      <h3>Edit connection</h3>
      <fieldset :disabled="changing" class="nh-edit-fields">
        <label class="nh-field" for="esp-edit-name">Name <input id="esp-edit-name" v-model="editName" class="nh-input" :maxlength="ESPHOME_MAX_NAME_LENGTH" /></label>
        <label class="nh-field" for="esp-edit-address">IP address <input id="esp-edit-address" v-model="editAddress" class="nh-input" required /></label>
        <label class="nh-field" for="esp-edit-port">Native API port <input id="esp-edit-port" v-model.number="editPort" class="nh-input" type="number" :min="ESPHOME_PORT_MIN" :max="ESPHOME_PORT_MAX" required /></label>
        <label class="nh-field" for="esp-edit-key-mode">Encryption <select id="esp-edit-key-mode" v-model="keyMode" class="nh-input"><option value="keep">Keep current setting</option><option value="replace">Replace key</option><option value="clear">Use unencrypted connection</option></select></label>
        <label v-if="keyMode === 'replace'" class="nh-field" for="esp-edit-key">New encryption key <input id="esp-edit-key" v-model="editKey" class="nh-input" type="password" autocomplete="off" /></label>
      </fieldset>
      <div>
        <UiButton :disabled="changing" @click="closeEdit">Cancel</UiButton>
        <UiButton type="submit" variant="primary" :disabled="changing" :loading="saving">Save connection</UiButton>
      </div>
    </form>
    <template #footer><UiButton @click="emit('close')">Close</UiButton></template>
  </UiDialog>
</template>

<style scoped>
  .nh-edit-fields {
    display: grid;
    gap: var(--nh-space);
    border: 0;
    padding: 0;
    margin: 0;
    min-width: 0;
  }
  .nh-provider-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--nh-space);
    padding-block: 0.5rem;
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
</style>
