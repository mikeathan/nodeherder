import mitt, { Emitter } from 'mitt';
import { onUnmounted } from 'vue';
TODO
export type OpenDialogEvent = {
  type: string; // e.g., 'confirm', 'nameDialog', etc.
  props?: Record<string, unknown>; // dynamic props for each dialog type
};

export type DialogEvents = {
  openDialog: OpenDialogEvent;
  closeDialog: void;
};

export type DialogEventHandlers = {
  [K in keyof DialogEvents]: (event: DialogEvents[K]) => void;
};

const dialogEventBus = mitt<DialogEvents>();

// Emitters
export function emitOpenDialog(event: OpenDialogEvent) {
  dialogEventBus.emit('openDialog', event);
}

export function emitCloseDialog() {
  dialogEventBus.emit('closeDialog');
}

// Composable for using dialog events
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
