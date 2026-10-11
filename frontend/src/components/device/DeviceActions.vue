<script setup lang="ts">
  /*
   * Bridge operations for a device: rename, interview, remove (with force/block options).
   * Offered only when the device's protocol supports them (domain/devices.ts resolveProtocol).
   */
  import { computed, ref } from 'vue';
  import { Device } from '@/types/device';
  import { store } from '@/store';
  import { resolveProtocol } from '@/domain/devices';
  import { confirm } from '@/composables/useConfirm';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiField from '@/components/ui/UiField.vue';

  const props = defineProps<{ device: Device; compact?: boolean }>();
  const supports = computed(() => resolveProtocol(props.device.connection_type).supports);

  const renaming = ref(false);
  const newName = ref('');
  const renameError = ref<string | null>(null);
  function openRename() {
    newName.value = props.device.friendly_name;
    renameError.value = null;
    renaming.value = true;
  }
  function rename() {
    const name = newName.value.trim();
    if (!name) {
      renameError.value = 'Enter a name.';
      return;
    }
    if (name !== props.device.friendly_name) store.dispatch('hub/renameDevice', { name: props.device.friendly_name, newName: name });
    renaming.value = false;
  }

  async function interview() {
    const ok = await confirm({
      title: `Interview ${props.device.friendly_name}?`,
      message: 'The bridge asks the device again what it supports. Battery devices may need to be woken up first.',
      confirmLabel: 'Interview',
    });
    if (ok) store.dispatch('hub/interviewDevice', { id: props.device.id });
  }

  const removing = ref(false);
  const force = ref(false);
  const block = ref(false);
  function openRemove() {
    force.value = false;
    block.value = false;
    removing.value = true;
  }
  function remove() {
    store.dispatch('hub/removeDevice', { id: props.device.id, force: force.value, block: block.value });
    removing.value = false;
  }
</script>

<template>
  <div class="nh-dev-actions">
    <UiButton v-if="supports.rename" :size="compact ? 'sm' : 'md'" variant="ghost" icon="edit" :label="`Rename ${device.friendly_name}`" @click="openRename" />
    <UiButton v-if="supports.interview" :size="compact ? 'sm' : 'md'" variant="ghost" icon="interview" :label="`Interview ${device.friendly_name}`" @click="interview" />
    <UiButton v-if="supports.remove" :size="compact ? 'sm' : 'md'" variant="ghost" icon="delete" class="is-danger" :label="`Remove ${device.friendly_name}`" @click="openRemove" />

    <UiDialog :open="renaming" title="Rename device" @close="renaming = false">
      <form :id="`rename-${device.id}`" @submit.prevent="rename">
        <UiField label="Name" :for="`rename-input-${device.id}`" :error="renameError" help="The bridge renames the device; automations keep working.">
          <input :id="`rename-input-${device.id}`" v-model="newName" class="nh-input" autocomplete="off" autofocus />
        </UiField>
      </form>
      <template #footer>
        <UiButton @click="renaming = false">Cancel</UiButton>
        <UiButton variant="primary" type="submit" :form="`rename-${device.id}`">Rename</UiButton>
      </template>
    </UiDialog>

    <UiDialog :open="removing" :title="`Remove ${device.friendly_name}?`" @close="removing = false">
      <p>The device leaves the network and disappears from NodeHerder. Areas and automations that use it will show it as removed.</p>
      <label class="nh-opt"><input v-model="force" type="checkbox" /> Force removal (when the device does not respond)</label>
      <label class="nh-opt"><input v-model="block" type="checkbox" /> Block it from joining again</label>
      <template #footer>
        <UiButton @click="removing = false">Cancel</UiButton>
        <UiButton variant="danger" icon="delete" @click="remove">Remove device</UiButton>
      </template>
    </UiDialog>
  </div>
</template>

<style scoped>
  .nh-dev-actions {
    display: inline-flex;
    gap: 0.1rem;
  }
  .nh-opt {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-height: 2.2rem;
    cursor: pointer;
  }
  .nh-opt input {
    accent-color: var(--nh-accent);
    width: 1.05rem;
    height: 1.05rem;
  }
</style>
