<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import InputBox from '../input/InputBox.vue';
import { store } from '../../store/index';
import { Device } from '@/types/device';
import {
  Automation,
  AutomationTrigger,
  TimeSchedule,
} from '@/types/automation';
import {
  EditableAutomationTrigger,
  DeviceAutomation,
} from '../../contracts/automations';
import AutomationStatus from '@/components/automations/schedule/AutomationStatus.vue';
import Panel from '../controls/Panel.vue';
import ButtonPanel from '@/components/controls/ButtonPanel.vue';
import { createEditAutomationButtonItems } from '../../configs/automation/trigger-dropdown.config';
import { emitCloseLastPanel } from '@/mixins/useAutomationsEventBus';
import {
  emitOpenSchedulerPanelEvent,
  emitOpenTriggerPanelEvent,
} from '@/contracts/panel-events';
import { DataTableRowClickEvent } from 'primevue';

const emit = defineEmits(['cancel']);

const props = defineProps({
  id: String,
});

const router = useRouter();
const automation = ref<Automation>({} as Automation);
const isInViewMode = ref<boolean>(true);
const buttonPanelItems = computed(() => {
  const isActionValid =
    automation.value.triggers.length == 0 &&
    automation.value.triggers.filter(
      (k) => k.action != null,
    ).length == automation.value.triggers.length;
  return createEditAutomationButtonItems(
    () => saveAutomation(),
    () => deleteAutomation(),
    () => openScheduler(automation.value),
    () => cancel(),
    isActionValid,
    isActionValid,
    isActionValid,
  );
});

watch(
  () => props.id,
  () => {
    var sourceAutomation = store.getters[
      'automations/find'
    ](props.id) as Device;
    if (sourceAutomation != undefined) {
      // make a deep copy to make it not reactive
      automation.value = JSON.parse(
        JSON.stringify(sourceAutomation),
      ) as DeviceAutomation;
    } else {
      automation.value = new DeviceAutomation();
      var device = store.getters['devices/find'](
        props.id,
      ) as Device;
      if (device != undefined) {
        automation.value.id = device.id;
        automation.value.friendlyname =
          device.friendly_name;
      }

      createNewTrigger();
    }
  },
  { immediate: true },
);

function createNewTrigger() {
  const newTrigger = EditableAutomationTrigger.create();

  emitOpenTriggerPanelEvent(
    automation.value.id,
    newTrigger,
    saveTrigger,
    deleteTrigger,
  );
}

function openScheduler(automation: Automation) {
  emitOpenSchedulerPanelEvent(automation);
}

function cancel() {
  emit('cancel');
}

function rowClicked(event: DataTableRowClickEvent): void {
  const trigger = automation.value.triggers[event.index];
  emitOpenTriggerPanelEvent(
    automation.value.id,
    trigger,
    saveTrigger,
    deleteTrigger,
  );
}

function onDeleteTriggerClick(
  event: Event,
  trgger: AutomationTrigger,
): void {
  deleteTrigger(trgger);
}

function onComponentDisplayed() {
  isInViewMode.value = false;
}
function onComponentHidden() {
  isInViewMode.value = true;
}
function createCloseLastPanelEvent() {
  emitCloseLastPanel();
}

function saveAutomation() {
  store.dispatch(
    'automations/save',
    automation.value as Automation,
  );
  router.push('/viewer');
}

function deleteAutomation() {
  var sourceAutomation = store.getters['automations/find'](
    props.id,
  );
  if (sourceAutomation != undefined) {
    store.dispatch(
      'automations/delete',
      automation.value.id,
    );
    // todo; alert message box to ask user
    router.push('/viewer');
  }
}

function deleteTrigger(trigger: AutomationTrigger): void {
  automation.value.triggers =
    automation.value.triggers.filter(
      (e, i) => e != trigger,
    );
}

function saveTrigger(trigger: AutomationTrigger): void {
  const idx = automation.value.triggers.indexOf(trigger);
  if (idx == -1) {
    automation.value.triggers.push(trigger);
  } else {
    automation.value.triggers[idx] = trigger;
  }
}

function getConditionsDescription(
  trigger: AutomationTrigger,
): string {
  var conditions = trigger.conditions;
  if (conditions.length == 0) {
    return '';
  }
  var condition = conditions[0];
  var description =
    condition.name +
    ' ' +
    condition.equality +
    ' ' +
    condition.value;
  if (conditions.length > 1) {
    description += '...';
  }

  return description;
}

function getActionDescription(
  trigger: AutomationTrigger,
): string {
  if (trigger.action?.id == '') {
    return '<EMPTY>';
  }
  return `${trigger.action.friendlyname}.${trigger.action.property}`;
}
</script>

<template>
  <!-- TODO: find better way to do this
    we have 2 components that use the same template and toggle from the if isinVieMode -->
    <div class="grid ">
    <Card
    v-bind:style="{
      display: isInViewMode ? 'block' : 'none',
    }"
    class="flex">
    <template #title>
      <div class="grid">
        <div class="row">
          <div
            class="col-12 xl:col-8 lg:col-8 sm:col-8 pb-3">
            <InputBox
              label="Id"
              :disabled="true"
              :value="automation.id"
              class="w-full" />
          </div>
          <div
            class="col-12 xl:col-8 lg:col-8 sm:col-8 pb-3">
            <InputBox
              label="Friendly Name"
              :disabled="true"
              :value="automation.friendlyname"
              class="w-full" />
          </div>
          <div
            class="col-12 xl:col-8 lg:col-8 sm:col-8 pb-3">
            <InputBox
              label="Description"
              @updated="(v) => (automation.description = v)"
              :value="automation.description"
              class="w-full" />
          </div>
          <div
            class="col-12 xl:col-8 lg:col-8 sm:col-8 pb-3">
            <AutomationStatus
              :automation="automation"
              :clickToOpen="true" />
          </div>
        </div>
      </div>
    </template>
    <template #content>
      <ButtonPanel :buttons="buttonPanelItems" />
      <div class="card">
        <DataTable 
          :value="automation.triggers"
          tableStyle="min-width: 50rem"
          @row-click="rowClicked"
          selectionMode="single">
          <Column field="action" header="Action">
            <template #body="slotProps">
              {{ getActionDescription(slotProps.data) }}
            </template>
          </Column>
          <Column field="conditions" header="Conditions">
            <template #body="slotProps">
              {{ getConditionsDescription(slotProps.data) }}
            </template>
          </Column>
          <Column>
            <template #header="slotProps">
              <Button
                icon="pi pi-plus"
                variant="text"
                rounded
                @click="createNewTrigger()" />
            </template>
            <template #body="slotProps">
              <Button
                icon="pi pi-trash"
                variant="text"
                rounded
                @click="
                  onDeleteTriggerClick(
                    $event,
                    slotProps.data,
                  )
                " />
            </template>
          </Column>
        </DataTable>
      </div>
      <!-- <table class="table responsive table-hover">
                <thead>
                    <tr>
                        <th scope="col">#</th>
                        <th scope="col">Action</th>
                        <th scope="col">Conditions</th>
                        <th scope="col">
                            <button type="button" class="btn btn-default btn-number" @click="createNewTrigger()">
                                <span class="fa fa-plus"></span>
                            </button>
                        </th>
                    </tr>
                </thead>
                <tbody v-for="(trigger, index) in automation.triggers" :item="trigger">
                    <tr>
                        <th scope="row">
                            {{ index + 1 }}
                        </th>
                        <td @click="rowClicked(trigger)">
                            {{ getActionDescription(trigger) }}
                        </td>
                        <td>
                            {{ getConditionsDescription(trigger) }}
                        </td>
                        <td>
                            <span class="fa fa-trash-alt fa-sm" @click="
                                onDeleteTriggerClick($event, trigger)
                                " data-bs-toggle="collapse" data-bs-target>
                            </span>
                        </td>
                    </tr>
                </tbody>
            </table> -->
    </template>
  </Card>
  <Card
    v-bind:style="{
      display: isInViewMode == false ? 'block' : 'none',
    }">
    <template #content>
      <Button
        icon="pi pi-times"
        variant="text"
        rounded
        class="float-end"
        @click="createCloseLastPanelEvent" />

      <Panel
        @close="onComponentHidden"
        @component-displayed="onComponentDisplayed">
      </Panel>
    </template>
  </Card>
  </div>
</template>
