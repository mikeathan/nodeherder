import { KeyValuePair } from './types';

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
