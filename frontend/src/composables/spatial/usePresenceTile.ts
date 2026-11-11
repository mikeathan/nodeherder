import { computed } from 'vue';
import type { Device } from '@/types/device';
import { derivedPresenceDisplay } from '@/domain/spatial/derived';
import {
  presenceDistanceLabel,
  presenceStateLabel,
  presenceTargetLabel,
  presenceZoneLabel,
} from '@/domain/spatial/presence';
import { useDerivedPresence } from './useDerivedPresence';
import { useSpatialPreview } from './useSpatialPreview';

/**
 * View model for one compact presence row. The component renders labels and preview
 * state; it does not derive presence or format text itself.
 */
export function usePresenceTile(device: () => Device, active: () => boolean) {
  const derived = useDerivedPresence(device);
  const preview = useSpatialPreview(() => device().id, active);
  const values = computed(() => derived.values.value);
  const stateLabel = computed(() => presenceStateLabel(derivedPresenceDisplay(values.value)));
  const targetLabel = computed(() => presenceTargetLabel(values.value.targetCount));
  const zoneLabel = computed(() => presenceZoneLabel(values.value.lastZone));
  const referenceLabel = computed(() => presenceDistanceLabel(preview.targetDistanceMm.value));
  return { preview, values, stateLabel, targetLabel, zoneLabel, referenceLabel };
}
