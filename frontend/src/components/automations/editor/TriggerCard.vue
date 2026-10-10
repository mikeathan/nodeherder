<script setup lang="ts">
  /*
   * One trigger as WHEN → IF → THEN (spec 007 AC-10…AC-13), with a plain-language summary,
   * per-field errors and warnings. Collapses to its summary.
   */
  import { computed, ref } from 'vue';
  import { Device } from '@/types/device';
  import { Automation, AutomationTrigger, ConditionType } from '@/types/automation.type';
  import { createActionFromType, createConditionFromType } from '@/contracts/automations';
  import { describeTrigger, Issue, issuesUnder } from '@/domain/automation';
  import { moveItem } from '@/domain/dashboard';
  import { findDevice } from '@/composables/useHub';
  import UiSelect from '@/components/ui/UiSelect.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import ConditionRow from './ConditionRow.vue';
  import ActionCard from './ActionCard.vue';
  import IssueText from './IssueText.vue';
  import { exposeOptions, notConfig } from './options';

  const props = defineProps<{
    automation: Automation;
    automationDevice: Device | undefined;
    path: string;
    errors: Issue[];
    warnings: Issue[];
    index: number;
    count: number;
    canRun: boolean;
  }>();
  defineEmits<{ (e: 'move', to: number): void; (e: 'remove'): void; (e: 'run'): void }>();
  const trigger = defineModel<AutomationTrigger>({ required: true });

  const open = ref(true);
  const bodyId = computed(() => `trigger-body-${props.index}`);
  const summary = computed(() => describeTrigger(props.automation, trigger.value, findDevice));
  const ownErrors = computed(() => issuesUnder(props.errors, props.path));
  const ownWarnings = computed(() => issuesUnder(props.warnings, props.path));

  const addCondition = (type: ConditionType) => trigger.value.conditions.push(createConditionFromType(type));
  const moveCondition = (from: number, to: number) => (trigger.value.conditions = moveItem(trigger.value.conditions, from, to));
  const addAction = () => trigger.value.actions.push(createActionFromType('trigger'));
  const moveAction = (from: number, to: number) => (trigger.value.actions = moveItem(trigger.value.actions, from, to));
</script>

<template>
  <section class="nh-trig" :class="{ 'has-err': ownErrors.length, 'is-collapsed': !open }" :aria-label="`Trigger ${index + 1}`">
    <header class="nh-trig-head">
      <button type="button" class="nh-trig-toggle" :aria-expanded="open" :aria-controls="bodyId" @click="open = !open">
        <UiIcon :name="open ? 'chevronDown' : 'chevronRight'" />
        <b>Trigger {{ index + 1 }}</b>
      </button>
      <span class="nh-trig-tools">
        <UiButton v-if="canRun" size="sm" icon="run" @click="$emit('run')">Run now</UiButton>
        <UiButton size="sm" variant="ghost" icon="moveUp" label="Move trigger up" :disabled="index === 0" @click="$emit('move', index - 1)" />
        <UiButton size="sm" variant="ghost" icon="moveDown" label="Move trigger down" :disabled="index === count - 1" @click="$emit('move', index + 1)" />
        <UiButton size="sm" variant="ghost" icon="delete" label="Delete trigger" @click="$emit('remove')" />
      </span>
      <p class="nh-sentence">{{ summary }}</p>
    </header>

    <div v-show="open" :id="bodyId" class="nh-flow">
      <div class="nh-block is-when">
        <div class="nh-block-label"><span>When</span></div>
        <div class="nh-block-body">
          <div class="nh-inline">
            <span class="nh-src"><UiIcon name="devices" />{{ automationDevice?.friendly_name ?? automation.friendlyname }}</span>
            <UiSelect v-model="trigger.name" :options="exposeOptions(automationDevice, notConfig, trigger.name)" label="Property that starts this trigger" placeholder="Choose what changes" :invalid="errors.some((e) => e.path === `${path}.name`)" />
            <span class="nh-muted">changes</span>
          </div>
          <IssueText :issues="errors" :path="`${path}.name`" />
        </div>
      </div>

      <div class="nh-block is-if">
        <div class="nh-block-label"><span>If</span></div>
        <div class="nh-block-body">
          <p v-if="!trigger.conditions.length" class="nh-muted nh-block-note">No conditions. Device-started runs need at least one; it can still be run by hand.</p>
          <ConditionRow
            v-for="(c, i) in trigger.conditions"
            :key="i"
            v-model="trigger.conditions[i]"
            :device="automationDevice"
            :path="`${path}.conditions.${i}`"
            :issues="errors"
            :index="i"
            :count="trigger.conditions.length"
            @move="moveCondition(i, $event)"
            @remove="trigger.conditions.splice(i, 1)" />
          <div class="nh-inline">
            <UiButton size="sm" icon="add" @click="addCondition('expose')">Value condition</UiButton>
            <UiButton size="sm" icon="clock" @click="addCondition('time')">Time window</UiButton>
          </div>
        </div>
      </div>

      <div class="nh-block is-then">
        <div class="nh-block-label"><span>Then</span></div>
        <div class="nh-block-body">
          <ActionCard
            v-for="(a, i) in trigger.actions"
            :key="i"
            v-model="trigger.actions[i]"
            :automation-device="automationDevice"
            :path="`${path}.actions.${i}`"
            :issues="errors"
            :index="i"
            :count="trigger.actions.length"
            @move="moveAction(i, $event)"
            @remove="trigger.actions.splice(i, 1)" />
          <IssueText :issues="errors" :path="`${path}.actions`" />
          <div class="nh-inline">
            <UiButton size="sm" icon="add" @click="addAction">Add action</UiButton>
          </div>
        </div>
      </div>
    </div>

    <ul v-if="ownWarnings.length" class="nh-warn-list">
      <li v-for="w in ownWarnings" :key="w.path + w.message"><UiIcon name="warn" />{{ w.message }}</li>
    </ul>
  </section>
</template>
