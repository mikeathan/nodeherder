import { defineAsyncComponent } from 'vue';

type DialogKey = string;
type Map = { [key: DialogKey]: any };

export const DialogComponents: Map = {
  confirm: defineAsyncComponent(() => import('../components/dialogs/ConfirmDialog.vue')),
  selection: defineAsyncComponent(() => import('../components/dialogs/SelectionDialog.vue')),
  exposeSelection: defineAsyncComponent(() => import('../components/dialogs/ExposeSelectionDialog.vue')),
  entityView: defineAsyncComponent(() => import('../components/dialogs/EntityViewDialog.vue')),
};
