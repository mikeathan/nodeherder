import { AutomationTrigger } from './automation.type';
import { KeyValuePair } from './types.type';

export type SaveTriggerFunc = (
  trigger: AutomationTrigger
) => void;
export type DeleteTriggerFunc = (
  trigger: AutomationTrigger
) => void;

export type EventAction = (...args: any) => void;
export type EventActions = KeyValuePair<EventAction>;

export type OpenPanelEvent = {
  name: string;

  args: any;
  events: EventActions;
};

export type Events = {
  openPanel: OpenPanelEvent;
  closePanel: string;
  closeLastPanel: void;
};
