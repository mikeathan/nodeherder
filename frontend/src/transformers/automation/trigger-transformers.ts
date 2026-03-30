import { AutomationConditionTypes, AutomationTrigger, isExposeCondition } from '@/types/automation.type';
import { escapeHTML } from '@/utils/html';

export function formatTriggerConditions(trigger: AutomationTrigger): string {
  const conditions = trigger.conditions;
  if (conditions.length == 0) {
    return '';
  }

  const condition = conditions[0];
  const hasTimeCondition = conditions.some((e) => e.type == AutomationConditionTypes.Time);
  let description = '';

  if (hasTimeCondition) {
    description += '<i class="pi pi-stopwatch pe-2" aria-label="Time Range Set"></i>';
  }
  if (isExposeCondition(condition)) {
    description += `${escapeHTML(condition.name)} ${escapeHTML(condition.equality)} ${escapeHTML(condition.value)}`;
    if (conditions.length > 1) {
      description += ` <span class="ml-1"><span >+${conditions.length - 1}</span></span>`;
    }
  }
  return description;
}
