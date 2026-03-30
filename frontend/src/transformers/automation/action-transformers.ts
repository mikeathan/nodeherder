import {
  AutomationTriggerActionExpose,
  AutomationTriggerAction,
  AutomationPresetCyclingAction,
  AutomationStepAction,
  AutomationActionStep,
  PublishModes,
} from '@/types/automation.type';
import { escapeHTML } from '@/utils/html';

function formatTriggerActionExposes(exposes: AutomationTriggerActionExpose[]): string {
  const listItems = exposes
    .map(
      (expose) =>
        `<li>Set <strong>${escapeHTML(expose.name)}</strong> to <strong>${escapeHTML(expose.data)}</strong></li>`
    )
    .join('');
  return `<ul>${listItems}</ul>`;
}

function formatTriggerActionOperations(triggerAction: AutomationTriggerAction): string {
  if (!triggerAction.delay) {
    return '';
  }
  return `Delay: ${escapeHTML(triggerAction.delay.value)} ${escapeHTML(triggerAction.delay.unit)}`;
}

export function transformTriggerAction(friendlyname: string, triggerAction: AutomationTriggerAction): string[] {
  const exposes = formatTriggerActionExposes(triggerAction.exposes);
  const action = `Trigger <strong>${escapeHTML(friendlyname)}</strong> ${exposes}`;
  const operations = formatTriggerActionOperations(triggerAction);

  if (triggerAction.publishMode == PublishModes.Single) {
    const broadcastMode = 'BroadcastMode: Single';
    return [action, operations, broadcastMode];
  }
  return [action, operations];
}

export function transformPresetCyclingAction(friendlyname: string, action: AutomationPresetCyclingAction): string[] {
  return [
    `Cycle <strong>${escapeHTML(action.property)}</strong> presets in <strong>${escapeHTML(friendlyname)}</strong>`,
  ];
}

function formatStepOperations(steps: AutomationActionStep[]): string {
  const hasPlus = steps.some((op) => op.operator === '+');
  if (hasPlus) {
    return 'Increase';
  }
  const hasMinus = steps.some((op) => op.operator === '-');
  if (hasMinus) {
    return 'Decrease';
  }
  return 'Unknown';
}

function findActionStepProperty(action: AutomationStepAction): string {
  return action.steps
    .map((step) => {
      if (step.id === action.id) {
        return step.property;
      }
    })
    .join('');
}

export function transformStepAction(friendlyname: string, action: AutomationStepAction): string[] {
  const operationType = formatStepOperations(action.steps);
  const expose = findActionStepProperty(action);

  // TODO add support for icons in operation
  //<i class="pi pi-plus" style="font-size: 0.5rem;"></i>
  const formattedAction = `${operationType} <strong>${escapeHTML(friendlyname)}</strong> <strong>${escapeHTML(
    expose
  )}</strong> by  ${escapeHTML(action.data)}`;
  return [formattedAction];
}
