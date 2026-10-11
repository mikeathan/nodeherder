<script setup lang="ts">
  /* Pick a device and the values to show as tiles in an area (spec 007 US-02). */
  import { computed, ref, watch } from 'vue';
  import { DashboardGroup } from '@/types/settings.type';
  import { useHub } from '@/composables/useHub';
  import { exposeLabel } from '@/domain/exposes';
  import { ExposeCategories } from '@/types/device.type';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';

  const props = defineProps<{ open: boolean; group: DashboardGroup | null }>();
  const emit = defineEmits<{ (e: 'close'): void; (e: 'add', deviceId: string, exposes: string[]): void }>();

  const { devices, findDevice } = useHub();
  const deviceId = ref<string | null>(null);
  const picked = ref<string[]>([]);

  watch(
    () => props.open,
    (open) => {
      if (!open) return;
      deviceId.value = null;
      picked.value = [];
    }
  );
  watch(deviceId, () => (picked.value = []));

  const deviceOptions = computed(() => devices.value.map((d) => ({ value: d.id, label: d.friendly_name })));
  const already = computed(() => (deviceId.value && props.group?.deviceGroup[deviceId.value]?.exposes) || []);
  const choices = computed(() => {
    const d = deviceId.value ? findDevice(deviceId.value) : undefined;
    if (!d) return [];
    return Object.values(d.exposes)
      .filter((e) => e.category !== ExposeCategories.Config)
      .map((e) => ({ name: e.name, label: exposeLabel(e.name), diagnostic: e.category === ExposeCategories.Diagnostic, added: already.value.includes(e.name) }))
      .sort((a, b) => Number(a.diagnostic) - Number(b.diagnostic) || a.label.localeCompare(b.label));
  });

  function add() {
    if (deviceId.value && picked.value.length) emit('add', deviceId.value, picked.value);
  }
</script>

<template>
  <UiDialog :open="open" :title="`Add tiles to ${group?.name ?? ''}`" wide @close="emit('close')">
    <div class="nh-field">
      <label class="nh-label" for="tile-device">Device</label>
      <UiSelect id="tile-device" v-model="deviceId" :options="deviceOptions" placeholder="Choose a device" />
    </div>
    <fieldset v-if="choices.length" class="nh-checks-wrap">
      <legend class="nh-label">Values to show</legend>
      <div class="nh-checks">
        <label v-for="c in choices" :key="c.name" class="nh-check">
          <input v-model="picked" type="checkbox" :value="c.name" :disabled="c.added" />
          <span>{{ c.label }}<small v-if="c.added"> · already shown</small><small v-else-if="c.diagnostic"> · diagnostic</small></span>
        </label>
      </div>
    </fieldset>
    <p v-else-if="deviceId" class="nh-muted">This device reports no values that can be shown as tiles.</p>
    <template #footer>
      <UiButton @click="emit('close')">Cancel</UiButton>
      <UiButton variant="primary" :disabled="!picked.length" @click="add">Add {{ picked.length || '' }} tile{{ picked.length === 1 ? '' : 's' }}</UiButton>
    </template>
  </UiDialog>
</template>

<style scoped>
  .nh-checks-wrap {
    border: 0;
    padding: 0;
    margin: 0;
    min-width: 0;
  }
  .nh-checks {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
    gap: 0.15rem 1rem;
    margin-top: 0.5rem;
  }
  .nh-check {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-height: 2.25rem;
    cursor: pointer;
  }
  .nh-check input {
    accent-color: var(--nh-accent);
    width: 1.1rem;
    height: 1.1rem;
  }
  .nh-check small {
    color: var(--nh-text-muted);
  }
</style>
