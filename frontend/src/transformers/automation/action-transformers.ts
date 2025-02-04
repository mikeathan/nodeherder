import {
  AutomationTriggerActionExpose,
  AutomationTriggerAction,
  AutomationPresetCyclingAction,
  AutomationStepAction,
} from '@/types/automation';

function formatTriggerActionExposes(exposes: AutomationTriggerActionExpose[]): string {
  const listItems = exposes
    .map((expose) => `<li>Set <strong>${expose.name}</strong> to <strong>${expose.data}</strong></li>`)
    .join('');
  return `<ul>${listItems}</ul>`;
}

function formatTriggerActionOperations(triggerAction: AutomationTriggerAction): string {
  if (!triggerAction.delay.value) {
    return '';
  }
  return `Delay: ${triggerAction.delay.value} ${triggerAction.delay.unit}`;
}

export function transformTriggerAction(friendlyname: string, triggerAction: AutomationTriggerAction): string[] {
  const exposes = formatTriggerActionExposes(triggerAction.exposes);
  const action = `Trigger <strong>${friendlyname}</strong> ${exposes}`;
  const operations = formatTriggerActionOperations(triggerAction);
  return [action, operations];
}

export function transformPresetCyclingAction(friendlyname: string, action: AutomationPresetCyclingAction): string[] {
  return [`Rotate <strong>${friendlyname}</strong> <ul><li>${action.property}</li></ul>`];
}

export function transformStepAction(friendlyname:string, action :AutomationStepAction):string[]{

  const formattedAction = `Adjust <strong>${friendlyname}</strong>`;
    //       currentAction.value.steps.forEach((step) => {
    //         stepValue +=
    //           step.property + ' ' + step.operator + ' ';
    //       });
    //       stepValue += currentAction.value.data;
    //       return [
    //         `Adjusting ${deviceNameFromId(currentAction.value)} ${currentAction.value.property}`,
    //         `by [${stepValue}] steps`,
    //       ];
    return []
}