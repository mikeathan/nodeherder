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
  RenameDeviceDialogProps,
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

export  function emitOpenRenameDeviceDialog(confirm: DialogEventAction, props: RenameDeviceDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'renameDevice',
    props: {
      show: true,
      ...props,
    },
    events: events,
  };
  emitOpenDialog(event);
}

export function emitOpenDeleteDeviceDialog(confirm: DeleteDeviceEventAction, props: DeleteDeviceDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'removeDevice',
    props: {
      show: true,
      ...props,
    },
    events: events,
  };
  emitOpenDialog(event);
}

export function emitOpenDeviceGroupSelectionDialog(confirm: DeviceGroupEventAction, props: DeviceGroupSelectionDialogProps) {
  const events: DialogEventActions = {
    close: () => emitCloseDialog(),
    confirm,
  };
  const event: OpenDialogEvent = {
    type: 'deviceGroupSelection',
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
