<script setup lang="ts">
  import { computed, reactive } from 'vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiSettingRow from '@/components/ui/UiSettingRow.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import { SCHEDULE_UNITS, formatScheduleInterval, toIntervalSeconds, type ScheduleUnit } from '@/domain/schedule';
  import { credentialLabel, credentialSlots } from '@/domain/integrations/credentials';
  import { formatLocalDateTime } from '@/utils/date.utils';
  import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations';
  const props = defineProps<{
    item: ConnectionInstance;
    definition: ProviderDefinition | null;
    busy: boolean;
    /**
     * Whether this provider reads state the hub already holds.
     *
     * Passed in rather than read from `definition`: for exactly these providers the definition is
     * deliberately not fetched, because there is no document to fetch. Deciding from it meant the
     * answer was always "not continuous", so this pane offered an address change and a refresh
     * schedule to a provider that has neither — and the schedule could only fail when saved.
     */
    continuous: boolean;
    /** Provider display name, resolved from the catalog because the definition may be absent. */
    providerName: string;
  }>();
  const name = defineModel<string>('name', { required: true });
  const scheduleEnabled = defineModel<boolean>('scheduleEnabled', { required: true });
  const scheduleValue = defineModel<number>('scheduleValue', { required: true });
  const scheduleUnit = defineModel<ScheduleUnit>('scheduleUnit', { required: true });
  const emit = defineEmits<{
    rename: [];
    toggleEnabled: [];
    editAddress: [];
    credential: [remove: boolean, slot?: string];
    credentials: [items: { slot: string; value: string }[]];
    saveSchedule: [];
    remove: [];
  }>();
  const tokens = reactive<Record<string, string>>({});
  const slots = computed(() => credentialSlots(props.definition?.auth));
  const hasCredentialInput = computed(() => slots.value.some((slot) => Boolean(tokens[slot])));
  const credentialsComplete = computed(() =>
    slots.value.length
      ? slots.value.every((slot) => props.item.credential_slots.includes(slot))
      : props.item.credential_configured
  );
  /** Two-way bridge for the toggle, which emits rather than writing a prop. */
  const observing = computed({
    get: () => props.item.enabled,
    set: (value: boolean) => {
      if (value !== props.item.enabled) emit('toggleEnabled');
    },
  });
  /**
   * What turning this off actually does.
   *
   * A continuous integration's switch stops the sensor being read, which is a data decision with a
   * cost on a shared device. Saying "disabled" alone would describe the flag rather than the effect.
   */
  const observingHelp = computed(() =>
    props.item.enabled
      ? 'On. Your sensor is being read and wireless devices are recorded.'
      : 'Off. Your sensor is not being read, so nothing is recorded.'
  );
  const observingHelpScheduled = computed(() =>
    props.item.enabled
      ? 'On. The provider is read on the schedule below, or when you refresh.'
      : 'Off. Nothing is fetched until you enable it again.'
  );
  function saveCredentials() {
    const items = slots.value.flatMap((slot) => {
      const value = tokens[slot];
      return value ? [{ slot, value }] : [];
    });
    if (!items.length) return;
    emit('credentials', items);
  }
  /** Clear only credentials the parent confirmed were saved; failed writes keep their input. */
  function clearSavedCredentials(slots: string[]) {
    for (const slot of slots) delete tokens[slot];
  }
  defineExpose({ clearSavedCredentials });
</script>
<template>
  <UiCard title="Integration settings"
    ><div class="connection-form">
      <div class="settings-provider">
        <span>Provider</span>
        <strong>{{ providerName }}</strong>
        <a
          v-if="definition && !continuous"
          :href="definition.base_url"
          target="_blank"
          rel="noopener noreferrer"
          >Open provider site ↗</a
        >
      </div>
      <!--
        Enable/disable is the primary control and its own row, not one of four peer buttons. An
        operator had to ask where it was, which is the whole argument: the only action that changes
        what the integration does was the least findable thing on the page.
      -->
      <UiSettingRow label="Observing" :help="continuous ? observingHelp : observingHelpScheduled">
        <UiToggle v-model="observing" label="Observing" :disabled="busy" :pending="busy" />
      </UiSettingRow>
      <UiSettingRow label="Integration name" for-id="settings-integration-name">
        <input id="settings-integration-name" v-model="name" maxlength="120" />
        <UiButton :disabled="busy || !name.trim()" @click="emit('rename')">Save name</UiButton>
      </UiSettingRow>
      <p v-if="continuous" class="connection-muted settings-inline-note">
        Reads every enrolled sensor directly, so it updates continuously. There is no address to enter
        and nothing to schedule.
      </p>
      <template v-else>
        <UiSettingRow label="Address" help="Changing the address uses a separate setup. Saved results remain available until the new address is verified.">
          <UiButton :disabled="busy" @click="emit('editAddress')">Change address</UiButton>
        </UiSettingRow>
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
      </template>
      <fieldset v-if="slots.length || item.credential_slots.length">
        <legend>Provider credentials</legend>
        <p class="connection-muted">
          {{
            credentialsComplete
              ? 'Credentials are saved. Their values cannot be displayed.'
              : 'Credentials are incomplete or missing.'
          }}
        </p>
        <p class="connection-muted">
          Saved slots not declared by the current configuration are unused. Removing a credential stops any pending
          attempt.
        </p>
        <label v-for="slot in slots" :key="slot" class="connection-field" :for="`settings-credential-${slot}`">
          <span>
            {{ credentialLabel(definition?.auth, slot) }}
            <em v-if="item.credential_slots.includes(slot)" class="settings-credential-saved">Saved</em>
          </span>
          <input
            :id="`settings-credential-${slot}`"
            v-model="tokens[slot]"
            type="password"
            autocomplete="off"
            maxlength="8192" />
        </label>
        <div class="connection-actions">
          <UiButton :disabled="busy || !hasCredentialInput" @click="saveCredentials">Save credentials</UiButton>
          <UiButton
            v-for="slot in item.credential_slots"
            :key="slot"
            :disabled="busy"
            @click="emit('credential', true, slot)"
            >Remove saved credential ({{ slot }})</UiButton
          >
        </div>
      </fieldset>
      <div class="connection-delete">
        <UiButton variant="danger" :disabled="busy" @click="emit('remove')">Remove integration</UiButton>
      </div>
    </div></UiCard
  >
</template>

<style scoped>
  .settings-provider {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 0.4rem 0.7rem;
    padding-bottom: 0.8rem;
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
  .settings-provider span {
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  .settings-provider a {
    font-size: 0.8rem;
    margin-inline-start: auto;
  }
  .settings-credential-saved {
    margin-inline-start: 0.4rem;
    color: var(--nh-ok);
    font-size: 0.7rem;
    font-style: normal;
    font-weight: 600;
  }
  /* A sentence where a control would be, so omitting it reads as a fact rather than a gap. */
  .settings-inline-note {
    margin: 0;
    max-width: 46ch;
  }
</style>
