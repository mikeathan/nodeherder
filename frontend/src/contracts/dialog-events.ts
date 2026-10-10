import { emitCloseDialog, emitOpenDialog, OpenDialogEvent } from '@/mixins/useDialogsEventBus';
import {
  BaseDialogProps,
  DeleteDeviceDialogProps,
  DeleteDeviceEventAction,
  DeviceGroupEventAction,
  DeviceGroupSelectionDialogProps,
  DialogEventAction,
  DialogEventActions,
  EntityViewDialogProps,
  ExposeSelectionDialogProps,
  InputDialogProps,
  RenameDeviceDialogProps,
  SelectionDialogProps,
} from '@/types/events.type';


export function emitOpenSelectionDialog(confirm: DialogEventAction, props: SelectionDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'selection',
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

export function emitOpenEntityViewDialog(confirm: DialogEventAction, props: EntityViewDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'entityView',
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
