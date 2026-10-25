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
  const mode = ref<'form' | 'yaml'>('form');
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

  function rememberFocus(event: FocusEvent) {
    const target = event.target;
    if (target instanceof HTMLElement && ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) && target.id) {
      focusedField.value = target.id;
    }
  }

  function remember() {
    if (!provider.value || !scope.isActive()) return;
    const value: ProviderDraft = {
      yaml: yaml.value,
      etag: etag.value,
      mode: mode.value,
      dirty: dirty.value,
      scroll: scroller()?.scrollTop ?? 0,
      focus: focusedField.value,
      document: documentValue.value ? cloneDefinitionValue(documentValue.value) : undefined,
    };
    scope.commit('draft', { key: key.value, value });
  }

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
  }

  function acceptValidation(value: DefinitionValidation) {
    synchronize(() => {
      yaml.value = value.yaml;
      documentValue.value = cloneDefinitionValue(value.document);
      validation.value = value;
    });
  }

  async function restore(saved: ProviderDraft) {
    synchronize(() => {
      yaml.value = saved.yaml;
      etag.value = saved.etag;
      mode.value = saved.mode;
      documentValue.value = saved.document ? cloneDefinitionValue(saved.document) : null;
      dirty.value = saved.dirty;
    });
    await nextTick();
    if (!scope.isActive()) return;
    frame = requestAnimationFrame(() => {
      if (!scope.isActive()) return;
      const field = saved.focus ? revealConnectionField(saved.focus) : null;
      field?.focus({ preventScroll: true });
      scroller()?.scrollTo({ top: saved.scroll });
    });
  }

  onBeforeRouteLeave(async (to) => {
    // An expired session must reach Login immediately; a dirty-draft dialog cannot block logout.
    if (!scope.isActive()) return true;
    remember();
    if (
      dirty.value &&
      to.name !== RouteName.ConnectionHelp &&
      !String(to.path).startsWith(`/connections/providers/${key.value}`)
    ) {
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
  };
}
export type ProviderDraftState = ReturnType<typeof useProviderDraft>;
