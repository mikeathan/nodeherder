import { emitOpenPanel } from '@/mixins/useAutomationsEventBus';
import { Automation, AutomationTrigger, TimeSchedule } from '@/types/automation.type';
import { DeleteTriggerFunc, EventActions, OpenPanelEvent, SaveTriggerFunc } from '@/types/events.type';

export function emitOpenSchedulerPanelEvent(automation: Automation): void {
  const events: EventActions = {
    save: (schedules: TimeSchedule[]) => {
      automation.schedules = schedules;
      automation.enabled = schedules.length == 0;
    },
  };

  const openPanelEvent: OpenPanelEvent = {
    name: 'Scheduler',
    args: { schedules: automation.schedules ?? [] },
    events: events,
  };

  emitOpenPanel(openPanelEvent);
}

export function emitOpenTriggerPanelEvent(
  automationId: string,
  trigger: AutomationTrigger,
  saveFunc: SaveTriggerFunc,
  deleteFunc: DeleteTriggerFunc
): void {
  const events: EventActions = {
    save: (e: AutomationTrigger) => {
      saveFunc(e);
    },
    delete: (e: AutomationTrigger) => {
      deleteFunc(e);
    },
  };

  const openPanelEvent: OpenPanelEvent = {
    name: 'Trigger',
    args: { id: automationId, trigger: trigger },
    events: events,
  };
  emitOpenPanel(openPanelEvent);
}
