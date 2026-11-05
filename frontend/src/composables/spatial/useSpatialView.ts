// Owns the spatial tab's drafts, saved settings and correlated save lifecycle.
// Components receive ready-to-render state; workspace edits live in useWorkspaceEditor.
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import type { Device, Expose } from '@/types/device';
import { exposeDisplayLabel } from '@/domain/exposes';
import {
  SPATIAL_DEFAULT_ORIENTATION,
  SPATIAL_DEFAULT_SCALE_MM,
  createSpatialMappings,
  mergeSpatialMappings,
  hasCompleteSpatialMapping,
  isSpatialCoordinate,
  isSpatialFlag,
  spatialMappingWarning,
  spatialRowOptions,
  stableSpatialInventory,
  validateSpatialMappings,
  type SpatialMapping,
  type SpatialRole,
} from '@/domain/spatial';
import {
  SPATIAL_SAVE_FAILED,
  SPATIAL_SAVE_NOTICE,
  copyWorkspace,
  validateWorkspace,
  workspaceDraftKey,
} from '@/domain/spatial/workspace';
import type { SpatialWorkspace } from '@/types/spatial';
import { saveConfirmedDeviceSettings } from '@/services/device-settings';
import { createDeviceConfigOverride } from '@/contracts/settings';
import { useDeviceSettings } from '../useDeviceSettings';
import { useHub } from '../useHub';
import { useSpatialPresence } from './useSpatialPresence';

/**
 * Owns the spatial tab's editable drafts, saved settings, selector options and
 * live targets. Components receive ready-to-render state and callbacks only.
 */
export function useSpatialView(device: () => Device) {
  const { isConnected } = useHub();
  const { settings, hasOverride } = useDeviceSettings(() => device().id);

  const mappings = ref<SpatialMapping[]>(createSpatialMappings());
  const mappingError = ref('');
  const workspace = ref<SpatialWorkspace>(copyWorkspace());
  const saveNotice = ref('');
  const saveError = ref('');
  const saving = ref(false);
  let saveController: AbortController | null = null;
  let loadedDeviceId: string | null = null;
  let hydratedDraft = '';
  onBeforeUnmount(() => saveController?.abort());
  const scale = ref(SPATIAL_DEFAULT_SCALE_MM);
  const orientation = ref(SPATIAL_DEFAULT_ORIENTATION);

  const configured = computed(() => hasCompleteSpatialMapping(settings.value?.spatial?.mappings ?? []));
  // Options AND warning/label callbacks must read stable metadata. Reading the
  // live device inside a callback subscribes the editor to every report and can
  // reset a native dropdown before its change event commits the pending choice.
  const inventory = computed<Expose[]>((previous) =>
    stableSpatialInventory(previous, Object.values(device().exposes ?? {}))
  );
  const editorEntities = computed(() => Object.fromEntries(inventory.value.map((entity) => [entity.name, entity])));
  const coordinates = computed(() => inventory.value.filter(isSpatialCoordinate));
  const flags = computed(() => inventory.value.filter(isSpatialFlag));
  const rowOptions = computed(() => spatialRowOptions(mappings.value, coordinates.value, flags.value));
  const canSave = computed(() => isConnected.value && !!settings.value && !saving.value);

  function spatialDraft() {
    return {
      mappings: mappings.value.map((mapping) => ({ ...mapping })),
      workspace: copyWorkspace(workspace.value),
      scale: scale.value,
      orientation: orientation.value,
    };
  }

  watch(
    [() => device().id, () => settings.value?.spatial],
    ([id, saved]) => {
      const incoming = {
        mappings: mergeSpatialMappings(saved?.mappings),
        workspace: copyWorkspace(saved?.workspace),
        scale: saved?.scale ?? SPATIAL_DEFAULT_SCALE_MM,
        orientation: saved?.orientation ?? SPATIAL_DEFAULT_ORIENTATION,
      };
      const incomingKey = workspaceDraftKey(incoming);
      const currentKey = workspaceDraftKey(spatialDraft());
      if (id === loadedDeviceId) {
        // Matching confirmed settings need no replacement: preserve selection and Undo.
        if (incomingKey === currentKey) {
          hydratedDraft = incomingKey;
          return;
        }
        // A remote config refresh must not discard edits made locally or while saving.
        if (currentKey !== hydratedDraft) return;
      } else {
        saveController?.abort();
        saveController = null;
        saving.value = false;
        saveNotice.value = '';
        saveError.value = '';
        mappingError.value = '';
        loadedDeviceId = id;
      }
      mappings.value = incoming.mappings;
      workspace.value = incoming.workspace;
      scale.value = incoming.scale;
      orientation.value = incoming.orientation;
      hydratedDraft = incomingKey;
    },
    { immediate: true }
  );
  watch(
    mappings,
    () => {
      mappingError.value = '';
    },
    { deep: true }
  );

  /** Applies one selector change without replacing the draft array. */
  function setMapping(row: number, role: SpatialRole, value: string) {
    mappings.value[row][role] = value;
  }

  /** Resolves an unavailable saved ID from stable editor metadata, falling back to the ID. */
  function selectedLabel(id: string): string {
    const entity = editorEntities.value[id];
    return entity ? exposeDisplayLabel(entity) : id;
  }

  /** Checks a draft against stable labels; live readings do not invalidate editor rendering. */
  function mappingWarning(mapping: SpatialMapping, index: number): string {
    return spatialMappingWarning(mapping, editorEntities.value, index);
  }

  /** Validates and submits a draft; only a correlated persisted acknowledgement publishes saved settings.
   * Failed/aborted saves retain drafts and undo; later edits survive an acknowledgement for an earlier draft.
   */
  async function saveDisplay() {
    if (!canSave.value) return;
    mappingError.value =
      validateSpatialMappings(mappings.value, device().exposes) || validateWorkspace(workspace.value);
    if (mappingError.value) return;
    const id = device().id;
    const controller = new AbortController();
    saveController = controller;
    saving.value = true;
    saveNotice.value = '';
    saveError.value = '';
    const base = hasOverride.value
      ? settings.value
      : {
          ...createDeviceConfigOverride(id),
          disabled: settings.value.disabled,
          metricsEnabled: settings.value.metricsEnabled,
        };
    try {
      await saveConfirmedDeviceSettings({ ...base, id, spatial: spatialDraft() }, controller.signal);
      if (saveController !== controller || controller.signal.aborted || device().id !== id) return;
      saveNotice.value = SPATIAL_SAVE_NOTICE;
    } catch (error: unknown) {
      if (saveController !== controller || controller.signal.aborted || device().id !== id) return;
      saveError.value = error instanceof Error ? error.message : SPATIAL_SAVE_FAILED;
    } finally {
      if (saveController === controller) {
        saving.value = false;
        saveController = null;
      }
    }
  }

  const {
    online,
    statuses: targetStatuses,
    tracks,
    slots,
    occupancy,
    summary,
    source,
    hasData,
    jitterMm,
    duplicates,
    replayFrames,
    recentTargets,
    smoothing,
  } = useSpatialPresence(
    () => device(),
    () => mappings.value,
    () => workspace.value
  );
  const showRaw = ref(false);

  return {
    mappings,
    workspace,
    saveNotice,
    saveError,
    saving,
    mappingError,
    scale,
    orientation,
    online,
    configured,
    rowOptions,
    canSave,
    tracks,
    slots,
    occupancy,
    summary,
    source,
    hasData,
    jitterMm,
    duplicates,
    replayFrames,
    recentTargets,
    showRaw,
    smoothing,
    targetStatuses,
    setMapping,
    selectedLabel,
    mappingWarning,
    saveDisplay,
  };
}
