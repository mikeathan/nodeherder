import { AutomationTrigger } from '@/types/automation.type';

export function formatTriggerConditions(trigger: AutomationTrigger): string {
  var conditions = trigger.conditions;
  if (conditions.length == 0) {
    return '';
  }

  var schedule = '';
  const hasSchedule = conditions.some((e) => e.timeRange);
  if (hasSchedule) {
    schedule += '<i class="pi pi-stopwatch pe-2" aria-label="Time Range Set"></i>';
  }

  var condition = conditions[0];
  var description = `${schedule}${condition.name} ${condition.equality} ${condition.value}`;
  if (conditions.length > 1) {
    description += ` <span class="ml-1"><span >+${conditions.length - 1}</span></span>`;
  }
  return description;
}
