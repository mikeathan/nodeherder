import { emitCloseDialog, emitOpenDialog, OpenDialogEvent } from '@/mixins/useDialogsEventBus';
import {
  BaseDialogProps,
  DialogEventAction,
  DialogEventActions,
  ExposeSelectionDialogProps,
} from '@/types/events.type';

export function emitOpenInputDialogEvent(confirm: DialogEventAction, props?: BaseDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };

  const event: OpenDialogEvent = {
    type: 'input',
    props: {
      show: true,
      ...props,
    },
    events: events,
  };

  emitOpenDialog(event);
}

export function emitOpenExposeSelectionDialog(confirm: DialogEventAction, props: ExposeSelectionDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'exposeSelection',
    props: {
      show: true,
      ...props,
    },
    events: events,
  };

  emitOpenDialog(event);
}

export function emitOpenDeviceSelectionDialog(confirm: DialogEventAction, props: BaseDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'deviceSelection',
    props: {
      show: true,
      ...props,
    },
    events: events,
  };

  emitOpenDialog(event);
}

export function emitOpenConfirmationDialog(confirm: DialogEventAction, props?: BaseDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'confirm',
    props: {
      show: true,
      ...props,
    },
    events: events,
  };
  emitOpenDialog(event);
}
