import { KeyValuePair } from '@/types/types.type';
import mitt, { Emitter } from 'mitt';
import { onUnmounted } from 'vue';

const DialogEventTypes = {
  confirm: 'confirm',
  input: 'input',
  renameDevice: 'renameDevice',
  removeDevice: 'removeDevice',
  exposeSelection: 'exposeSelection',
} as const;

export type CloseDialogFunc = (
) => void;

export type DialogEventAction = (...args: any) => void;
export type DialogEventActions = KeyValuePair<DialogEventAction>;

export type DialogEventType = keyof typeof DialogEventTypes;

export type OpenDialogEvent = {
  type: DialogEventType;
  props?: Record<string, any>;
  events: DialogEventActions;

};

export type DialogEvents = {
  openDialog: OpenDialogEvent;
  closeDialog: void;
};

export type DialogEventHandlers = {
  [K in keyof DialogEvents]: (event: DialogEvents[K]) => void;
};

export type DialogHandler<T = unknown> = (event: T) => void;

const dialogEventBus = mitt<DialogEvents>();

export function emitOpenDialog(event: OpenDialogEvent) {
  dialogEventBus.emit('openDialog', event);
}

export function emitCloseDialog() {
  dialogEventBus.emit('closeDialog');
}

export function useDialogEvents(handlers: DialogEventHandlers) {
  const keys = Object.keys(handlers) as Array<keyof DialogEvents>;
  for (const key of keys) {
    dialogEventBus.on(key, handlers[key] as never);
  }

  const cleanup = () => {
    for (const key of keys) {
      dialogEventBus.off(key, handlers[key] as never);
    }
  };

  onUnmounted(cleanup);
  return cleanup;
}

export default dialogEventBus;
