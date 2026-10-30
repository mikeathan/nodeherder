<script setup lang="ts">
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { SCHEDULE_UNITS, formatScheduleInterval, toIntervalSeconds, type ScheduleUnit } from '@/domain/schedule';
  import { formatLocalDateTime } from '@/utils/date.utils';
  import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations.type';
  defineProps<{ item: ConnectionInstance; definition: ProviderDefinition | null; busy: boolean }>();
  const name = defineModel<string>('name', { required: true });
  const token = defineModel<string>('token', { required: true });
  const scheduleEnabled = defineModel<boolean>('scheduleEnabled', { required: true });
  const scheduleValue = defineModel<number>('scheduleValue', { required: true });
  const scheduleUnit = defineModel<ScheduleUnit>('scheduleUnit', { required: true });
  const emit = defineEmits<{
    update: [enabled: boolean];
    editAddress: [];
    credential: [remove: boolean, slot?: string];
    saveSchedule: [];
    remove: [];
  }>();
</script>
<template>
  <UiCard title="Integration settings"
    ><div class="connection-form">
      <label class="connection-field"><span>Integration name</span><input v-model="name" maxlength="120" /></label>
      <div class="connection-actions">
        <UiButton :disabled="busy || !name.trim()" @click="emit('update', item.enabled)">Save name</UiButton
        ><UiButton :disabled="busy" @click="emit('update', !item.enabled)">{{
          item.enabled ? 'Disable integration' : 'Enable integration'
        }}</UiButton
        ><UiButton :disabled="busy" @click="emit('editAddress')">Change address</UiButton>
      </div>
      <p class="connection-muted">
        Changing the address uses a separate setup. Your saved dates remain available until the new address is verified.
      </p>
      <fieldset>
        <legend>Automatic refresh</legend>
        <label><input v-model="scheduleEnabled" type="checkbox" /> Refresh this integration automatically</label>
        <div class="connection-row">
          <label class="connection-field"
            ><span>Every</span
            ><input
              v-model.number="scheduleValue"
              type="number"
              :min="scheduleUnit === 'seconds' ? 60 : 1"
              step="1"
              :disabled="!scheduleEnabled"
          /></label>
          <label class="connection-field"
            ><span>Unit</span
            ><select v-model="scheduleUnit" :disabled="!scheduleEnabled">
              <option v-for="unit in SCHEDULE_UNITS" :key="unit.value" :value="unit.value">{{ unit.label }}</option>
            </select></label
          >
        </div>
        <p class="connection-muted">
          {{
            scheduleEnabled
              ? `Refreshes ${formatScheduleInterval(toIntervalSeconds(scheduleValue, scheduleUnit))}.`
              : 'Manual refresh only.'
          }}
          Saving this setting does not contact the provider.
        </p>
        <p v-if="item.schedule?.next_due_at" class="connection-muted">
          Next automatic refresh: {{ formatLocalDateTime(item.schedule.next_due_at) }}.
        </p>
        <p v-if="item.schedule?.last_attempt_at" class="connection-muted">
          Last attempt: {{ formatLocalDateTime(item.schedule.last_attempt_at) }}.
        </p>
        <p v-if="item.schedule?.last_success_at" class="connection-muted">
          Last successful refresh: {{ formatLocalDateTime(item.schedule.last_success_at) }}.
        </p>
        <p v-if="item.schedule?.last_failure_at" class="connection-muted">
          Last refresh failure: {{ formatLocalDateTime(item.schedule.last_failure_at) }} ({{
            item.schedule.consecutive_failures
          }}
          consecutive).
        </p>
        <div class="connection-actions">
          <UiButton :disabled="busy || (scheduleEnabled && !scheduleValue)" @click="emit('saveSchedule')"
            >Save schedule</UiButton
          >
        </div>
      </fieldset>
      <fieldset v-if="definition?.auth.secret_slot || item.credential_slots.length">
        <legend>Provider token</legend>
        <p class="connection-muted">
          {{ item.credential_configured ? 'A token is saved. Its value cannot be displayed.' : 'No token is saved.' }}
        </p>
        <p class="connection-muted">
          Saved tokens for other slot names are not used by the current configuration. Removing a token stops any
          pending attempt.
        </p>
        <label v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot" class="connection-field"
          ><span>{{
            item.credential_slots.includes(definition.auth.secret_slot) ? 'Replacement API token' : 'API token'
          }}</span
          ><input v-model="token" type="password" autocomplete="off" maxlength="8192"
        /></label>
        <div class="connection-actions">
          <UiButton
            v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot"
            :disabled="busy || !token"
            @click="emit('credential', false)"
            >Save token</UiButton
          ><UiButton
            v-for="slot in item.credential_slots"
            :key="slot"
            :disabled="busy"
            @click="emit('credential', true, slot)"
            >Remove saved token ({{ slot }})</UiButton
          >
        </div>
      </fieldset>
      <div class="connection-delete">
        <UiButton variant="danger" :disabled="busy" @click="emit('remove')">Remove integration</UiButton>
      </div>
    </div></UiCard
  >
</template>
