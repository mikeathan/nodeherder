<script setup lang="ts">
  /*
   * Explicit firmware-setting mapping and writes. Nothing is inferred from entity names: the
   * user maps each role, enters a value in the entity's own unit, and sees requested versus
   * reported state. Zone off delay is mapped per zone in the zone editor.
   */
  import { computed, reactive, ref, watch } from 'vue';
  import type { Device } from '@/types/device';
  import type { FirmwareRole, SpatialWorkspace } from '@/types/spatial';
  import {
    FIRMWARE_SETTINGS,
    ZONE_APPLY_MESSAGES,
    firmwareCommandPlan,
    firmwareEntityOptions,
    zoneCommandRows,
  } from '@/domain/spatial/device-zones';
  import { exposeDisplayLabel, formatExposeValue } from '@/domain/exposes';
  import { useDeviceCommand } from '@/composables/useDeviceCommand';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ device: Device; workspace: SpatialWorkspace; online: boolean }>();
  const emit = defineEmits<{ map: [role: FirmwareRole, id: string] }>();
  const { send, isPending, canControl } = useDeviceCommand();
  const values = reactive<Partial<Record<FirmwareRole, number | boolean>>>({});
  const notice = ref('');

  const mapped = computed(() =>
    FIRMWARE_SETTINGS.map((setting) => ({
      ...setting,
      entity: props.device.exposes[props.workspace.firmwareEntities?.[setting.role] ?? ''],
      options: firmwareEntityOptions(props.device.exposes, setting.role),
    }))
  );
  const plan = computed(() => firmwareCommandPlan(props.workspace, props.device.exposes, values));
  const rows = computed(() => zoneCommandRows(plan.value.commands, props.device.exposes));
  const busy = computed(() => plan.value.commands.some((command) => isPending(props.device.id, command.entity)));

  // Seed inputs from what the device reports whenever a mapping or reading changes.
  watch(
    () => mapped.value.map((setting) => [setting.role, setting.entity?.name, setting.entity?.data]),
    () => {
      for (const setting of mapped.value) {
        if (!setting.entity) continue;
        if (setting.kind === 'switch' && typeof setting.entity.data === 'boolean') values[setting.role] = setting.entity.data;
        if (setting.kind === 'number' && typeof setting.entity.data === 'number') values[setting.role] = setting.entity.data;
      }
    },
    { immediate: true, deep: true }
  );
  watch(plan, () => { notice.value = ''; });

  function apply() {
    if (plan.value.error || !props.online || !canControl(props.device) || busy.value) return;
    for (const command of plan.value.commands) {
      if (!send(props.device, command.entity, command.value)) {
        notice.value = ZONE_APPLY_MESSAGES.partial;
        return;
      }
    }
    notice.value = ZONE_APPLY_MESSAGES.requested;
  }
</script>

<template>
  <section aria-label="Firmware settings">
    <h3>Firmware settings</h3>
    <p class="nh-muted">
      Map only what this firmware exposes. Writes are individual, not atomic; a submitted value is not proof of
      persistence. Values use each entity’s own unit.
    </p>
    <div v-for="setting in mapped" :key="setting.role" class="nh-firmware-row">
      <label>
        {{ setting.label }} entity
        <select
          class="nh-input"
          :value="workspace.firmwareEntities?.[setting.role] ?? ''"
          @change="emit('map', setting.role, ($event.target as HTMLSelectElement).value)">
          <option value="">Not mapped</option>
          <option
            v-if="workspace.firmwareEntities?.[setting.role] && !setting.entity"
            :value="workspace.firmwareEntities[setting.role]"
            disabled>
            Unavailable: {{ workspace.firmwareEntities[setting.role] }}
          </option>
          <option v-for="option in setting.options" :key="option.name" :value="option.name">
            {{ exposeDisplayLabel(option) }}
          </option>
        </select>
      </label>
      <template v-if="setting.entity">
        <label v-if="setting.kind === 'switch'" class="nh-firmware-toggle">
          <input
            type="checkbox"
            :checked="values[setting.role] === true"
            @change="values[setting.role] = ($event.target as HTMLInputElement).checked" />
          On
        </label>
        <label v-else>
          Value ({{ setting.entity.unit || 'native' }})
          <input
            class="nh-input is-num"
            type="number"
            :value="values[setting.role] ?? ''"
            :step="setting.entity.attributes?.step ?? 'any'"
            @input="values[setting.role] = Number(($event.target as HTMLInputElement).value)" />
        </label>
        <span class="nh-muted">Reported: {{ formatExposeValue(setting.entity, setting.entity.data) }}</span>
      </template>
    </div>
    <ul v-if="!plan.error && rows.length">
      <li v-for="row in rows" :key="row.role">
        {{ row.role }}: request {{ row.request }} · reported {{ row.reported }}
        {{ isPending(device.id, row.entity) ? '(awaiting confirmation)' : '' }}
      </li>
    </ul>
    <p v-else-if="plan.error" class="nh-muted">{{ plan.error }}</p>
    <UiButton :disabled="!!plan.error || !plan.commands.length || !online || !canControl(device) || busy" @click="apply">
      Apply firmware settings
    </UiButton>
    <p v-if="notice" role="status">{{ notice }}</p>
  </section>
</template>

<style scoped>
  .nh-firmware-row {
    display: grid;
    gap: 0.4rem;
    padding-block: 0.5rem;
    border-block-end: 1px solid var(--nh-border);
  }
  .nh-firmware-row label {
    display: grid;
    gap: 0.3rem;
  }
  .nh-firmware-toggle {
    display: inline-flex;
    gap: 0.35rem;
    align-items: center;
  }
</style>
