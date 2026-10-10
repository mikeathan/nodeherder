<script setup lang="ts">
  /*
   * On/off for the automation: either a manual switch, or daily enable/disable times
   * (TimeSchedule, at most one of each type, as the previous editor allowed).
   */
  import { computed } from 'vue';
  import { Automation, TimeScheduleType } from '@/types/automation.type';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const automation = defineModel<Automation>({ required: true });
  const hasType = (type: TimeScheduleType) => automation.value.schedules?.some((s) => s.type === type) ?? false;
  const scheduled = computed(() => (automation.value.schedules?.length ?? 0) > 0);

  function add(type: TimeScheduleType) {
    automation.value.schedules = [...(automation.value.schedules ?? []), { type, startAt: type === 'enable' ? '07:00' : '23:00' }];
  }
  function remove(index: number) {
    automation.value.schedules = automation.value.schedules.filter((_, i) => i !== index);
  }
</script>

<template>
  <UiCard title="On or off" icon="schedule">
    <div class="nh-sched-switch">
      <span>
        <b>{{ automation.enabled ? 'On' : 'Off' }}</b>
        <small class="nh-muted">{{ scheduled ? ' · the schedule below switches it' : ' · runs whenever its triggers fire' }}</small>
      </span>
      <UiToggle v-model="automation.enabled" label="Automation on" />
    </div>
    <div v-for="(s, i) in automation.schedules ?? []" :key="s.type" class="nh-sched-row">
      <span class="nh-sched-t" :class="`is-${s.type}`">{{ s.type === 'enable' ? 'Turn on at' : 'Turn off at' }}</span>
      <input v-model="s.startAt" type="time" class="nh-input" :aria-label="s.type === 'enable' ? 'Turn on at' : 'Turn off at'" />
      <UiButton size="sm" variant="ghost" icon="delete" :label="`Remove ${s.type} time`" @click="remove(i)" />
    </div>
    <div class="nh-inline nh-sched-add">
      <UiButton v-if="!hasType('enable')" size="sm" icon="add" @click="add('enable')">Turn on daily at…</UiButton>
      <UiButton v-if="!hasType('disable')" size="sm" icon="add" @click="add('disable')">Turn off daily at…</UiButton>
    </div>
  </UiCard>
</template>

<style scoped>
  .nh-sched-switch {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }
  .nh-sched-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.6rem;
  }
  .nh-sched-t {
    flex: 1;
    font-size: 0.875rem;
    font-weight: 600;
  }
  .nh-sched-t::before {
    content: '';
    display: inline-block;
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    margin-right: 0.4rem;
    background: var(--nh-ok);
  }
  .nh-sched-t.is-disable::before {
    background: var(--nh-danger);
  }
  .nh-sched-add {
    margin-top: 0.75rem;
  }
</style>
