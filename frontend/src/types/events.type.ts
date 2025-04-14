import { AutomationTrigger } from './automation.type';
import { RemoveDeviceEvent } from './dialog.type';
import { KeyValuePair } from './types.type';

// Automation panel events
export type SaveTriggerFunc = (trigger: AutomationTrigger) => void;
export type DeleteTriggerFunc = (trigger: AutomationTrigger) => void;

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

// Dialog events
export const DialogEventTypes = {
  confirm: 'confirm',
  input: 'input',
  renameDevice: 'renameDevice',
  removeDevice: 'removeDevice',
  exposeSelection: 'exposeSelection',
  deviceSelection: 'deviceSelection',
} as const;

export type CloseDialogFunc = () => void;
export type BaseDialogProps = {
  title?: string;
  message?: string;
};

export type ExposeSelectionDialogProps = {
  id: string;
  title?: string;
  message?: string;
};

export type RenameDeviceDialogProps = {
  friendlyName: string;
  title?: string;
  message?: string;
};

export type DeleteDeviceDialogProps = {
  friendlyName: string;
  title?: string;
  message?: string;
};

export type DialogEventAction = (...args: any) => void;
export type DeleteDeviceEventAction = (args: RemoveDeviceEvent) => void;
export type DialogEventActions = KeyValuePair<DialogEventAction>;
