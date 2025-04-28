import { defineAsyncComponent } from 'vue';

type DialogKey = string;
type Map = { [key: DialogKey]: any };

export const DialogComponents: Map = {
  confirm: defineAsyncComponent(() => import('../components/dialogs/ConfirmDialog.vue')),
  input: defineAsyncComponent(() => import('../components/dialogs/InputDialog.vue')),
  renameDevice: defineAsyncComponent(() => import('../components/dialogs/RenameDeviceDialog.vue')),
  removeDevice: defineAsyncComponent(() => import('../components/dialogs/RemoveDeviceDialog.vue')),
  exposeSelection: defineAsyncComponent(() => import('../components/dialogs/ExposeSelectionDialog.vue')),
  deviceSelection: defineAsyncComponent(() => import('../components/dialogs/DeviceSelectionDialog.vue')),
  entityView: defineAsyncComponent(() => import('../components/dialogs/EntityViewDialog.vue')),
};
