import type { ProviderEditorMode } from '@/types/integrations.type';
import { nextTick, onBeforeUnmount, ref, shallowRef, watch, type Ref } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { confirm } from '@/composables/useConfirm';
import type { ConnectionScope } from '@/composables/useConnections';
import type {
  ConnectionDefinition,
  DefinitionValidation,
  ProviderDefinition,
  ProviderDraft,
} from '@/types/integrations.type';
import { revealConnectionField } from '@/components/integrations/field-focus';
import { RouteName } from '@/types/router';
import { cloneDefinitionValue } from './useEditableModel';

/** Owns the unsaved document and Help navigation. HTTP and provider execution live elsewhere. */
export function useProviderDraft(scope: ConnectionScope, key: Ref<string>) {
  const provider = shallowRef<ConnectionDefinition | null>(null);
  const yaml = ref('');
  const documentValue = shallowRef<ProviderDefinition | null>(null);
  const mode = ref<ProviderEditorMode>('form');
  const dirty = ref(false);
  const etag = ref('');
  const validation = shallowRef<DefinitionValidation | null>(null);
  const approve = ref(false);
  const reassociate = ref<string[]>([]);
  const focusedField = ref('');
  let hydrating = false;
  let frame: number | undefined;
  const scroller = () => document.querySelector<HTMLElement>('.nh-app-main');

  // Form sections emit replacements. A shallow watch avoids walking/cloning the complete document
  // on every keystroke; snapshots are taken only at navigation and explicit synchronization points.
  watch(
    [yaml, documentValue],
    () => {
      if (hydrating) return;
      dirty.value = true;
      validation.value = null;
      approve.value = false;
      reassociate.value = [];
    },
    { flush: 'sync' }
  );

  function synchronize(action: () => void) {
    hydrating = true;
    try {
      action();
    } finally {
      hydrating = false;
    }
  }

  /** Track the last identified form field without retaining DOM nodes after navigation. */
  function rememberFocus(event: FocusEvent) {
    const target = event.target;
    if (target instanceof HTMLElement && ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) && target.id) {
      focusedField.value = target.id;
    }
  }

  /** Save a session-only draft snapshot; an explicit scroll offset records section-tab intent. */
  function remember(scrollTop = scroller()?.scrollTop ?? 0) {
    if (!provider.value || !scope.isActive()) return;
    const value: ProviderDraft = {
      yaml: yaml.value,
      etag: etag.value,
      mode: mode.value,
      dirty: dirty.value,
      scroll: scrollTop,
      focus: focusedField.value,
      document: documentValue.value ? cloneDefinitionValue(documentValue.value) : undefined,
    };
    scope.commit('draft', { key: key.value, value });
  }

  /** Install an accepted definition and a separately owned editable document, clearing validation. */
  function apply(value: ConnectionDefinition) {
    synchronize(() => {
      provider.value = value;
      yaml.value = value.yaml;
      etag.value = value.etag;
      documentValue.value = value.status === 'active' ? cloneDefinitionValue(value.document) : null;
      mode.value = documentValue.value ? 'form' : 'yaml';
      dirty.value = false;
      validation.value = null;
      approve.value = false;
      reassociate.value = [];
    });
    scope.commit('definition', value);
  }

  /** Accept normalized editor values without marking synchronization as another user edit. */
  function acceptValidation(value: DefinitionValidation) {
    synchronize(() => {
      yaml.value = value.yaml;
      documentValue.value = cloneDefinitionValue(value.document);
      validation.value = value;
    });
  }

  /** Hydrate an owned draft before first render; keep its original edit token for conflict checks. */
  function hydrate(saved: ProviderDraft) {
    synchronize(() => {
      yaml.value = saved.yaml;
      etag.value = saved.etag;
      mode.value = saved.mode;
      documentValue.value = saved.document ? cloneDefinitionValue(saved.document) : null;
      dirty.value = saved.dirty;
    });
  }

  /** Restore the buffer and guarded focus/scroll after its fields have rendered. */
  async function restore(saved: ProviderDraft) {
    hydrate(saved);
    await nextTick();
    if (!scope.isActive()) return;
    if (frame !== undefined) cancelAnimationFrame(frame);
    frame = requestAnimationFrame(() => {
      frame = undefined;
      if (!scope.isActive()) return;
      const field = saved.focus ? revealConnectionField(saved.focus) : null;
      field?.focus({ preventScroll: true });
      scroller()?.scrollTo({ top: saved.scroll });
    });
  }

  onBeforeRouteLeave(async (to) => {
    // An expired session must reach Login immediately; a dirty-draft dialog cannot block logout.
    if (!scope.isActive()) return true;
    const sameProviderSection =
      [RouteName.IntegrationProvider, RouteName.IntegrationHelp].includes(to.name as RouteName) &&
      `${to.params.type}/${to.params.provider}` === key.value;
    // Top-level tabs return to the top. Contextual Help anchors still restore the original field
    // position; restoring a previous scroll offset after a tab click would visibly move the header.
    const sectionHelp = sameProviderSection && to.name === RouteName.IntegrationHelp && !to.hash;
    remember(sectionHelp ? 0 : undefined);
    if (dirty.value && !sameProviderSection) {
      return confirm({
        title: 'Leave provider editor?',
        message: 'Your draft is kept in memory while you stay signed in. It has not been saved to the host.',
        confirmLabel: 'Keep draft and leave',
      });
    }
    return true;
  });
  onBeforeUnmount(() => {
    if (frame !== undefined) cancelAnimationFrame(frame);
  });

  return {
    provider,
    yaml,
    documentValue,
    mode,
    dirty,
    etag,
    validation,
    approve,
    reassociate,
    rememberFocus,
    remember,
    apply,
    acceptValidation,
    restore,
    hydrate,
  };
}
export type ProviderDraftState = ReturnType<typeof useProviderDraft>;
