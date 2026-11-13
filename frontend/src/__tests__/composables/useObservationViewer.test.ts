import { describe, expect, test } from '@jest/globals';
import { ref } from 'vue';
import { useObservationViewer, OBSERVATION_VIEW_LABELS } from '@/composables/useObservationViewer';
import type { ObservationSnapshot } from '@/types/observations.type';
import {
  ObservationDecisions,
  ObservationTrends,
  ObservationViews,
} from '@/types/observations.type';

/** Minimal identity fixture; only the fields the viewer reads are meaningful. */
function identity(key: string, overrides: Record<string, unknown> = {}) {
  return {
    key,
    kind: 'address',
    stable: true,
    name: '',
    vendor: 0x004c,
    address: key.replace('addr:', ''),
    firstSeen: new Date().toISOString(),
    lastSeen: new Date().toISOString(),
    count: 5,
    rssi: -60,
    rssiSmoothed: -60,
    trend: ObservationTrends.Steady,
    decision: ObservationDecisions.Unknown,
    ...overrides,
  } as never;
}

/** Snapshot with the given identities and one source declaring the standard facets. */
function snapshot(identities: unknown[], facets = ['kind', 'vendor']): ObservationSnapshot {
  return {
    identities,
    sources: [
      {
        id: 'esp-1',
        name: 'Kitchen proxy',
        media: ['ble.advertisement'],
        facets,
        diagnostics: { advertisements: 10, decoded: 10, dropped: 0 },
      },
    ],
    truncated: false,
  } as never;
}

describe('useObservationViewer', () => {
  test('renders the live snapshot when not paused', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);

    expect(viewer.visible.value).toHaveLength(1);
    expect(viewer.paused.value).toBe(false);
  });

  test('groups facets declared by sources, not from a hardcoded list', () => {
    const state = ref(snapshot([], ['kind', 'addressType']));
    const viewer = useObservationViewer(() => state.value);

    expect(viewer.facets.value.sort()).toEqual(['addressType', 'kind']);
  });

  test('pausing freezes the rendered identities', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);

    viewer.togglePause();
    state.value = snapshot([
      identity('addr:aa:bb:cc:dd:ee:01'),
      identity('addr:aa:bb:cc:dd:ee:02'),
    ]);

    expect(viewer.visible.value).toHaveLength(1);
    expect(viewer.paused.value).toBe(true);
  });

  test('resuming follows the live snapshot again', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);

    viewer.togglePause();
    state.value = snapshot([
      identity('addr:aa:bb:cc:dd:ee:01'),
      identity('addr:aa:bb:cc:dd:ee:02'),
    ]);
    viewer.togglePause();

    expect(viewer.visible.value).toHaveLength(2);
  });

  test('a decision applied while paused is visible without resuming', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);

    viewer.togglePause();
    viewer.applyDecisionToFrozen('addr:aa:bb:cc:dd:ee:01', ObservationDecisions.Known);

    expect(viewer.visible.value[0]?.decision).toBe(ObservationDecisions.Known);
  });

  test('a decision outside the frozen view changes nothing', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);

    // Not paused: the frozen list does not exist, so this must be a no-op rather than a crash.
    expect(() => viewer.applyDecisionToFrozen('missing', ObservationDecisions.Known)).not.toThrow();
  });

  test('selecting an identity isolates it and switches to focus', () => {
    const state = ref(
      snapshot([identity('addr:aa:bb:cc:dd:ee:01'), identity('addr:aa:bb:cc:dd:ee:02')]),
    );
    const viewer = useObservationViewer(() => state.value);

    viewer.selectIdentity('addr:aa:bb:cc:dd:ee:02');

    expect(viewer.scoped.value).toHaveLength(1);
    expect(viewer.viewMode.value).toBe(ObservationViews.Focus);
    expect(viewer.availableViews.value).toContain(ObservationViews.Focus);
  });

  test('clearing a selection leaves focus', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);
    viewer.selectIdentity('addr:aa:bb:cc:dd:ee:01');

    viewer.clearSelection();

    expect(viewer.selected.value).toBeNull();
    expect(viewer.viewMode.value).toBe(ObservationViews.Radar);
  });

  test('an identity that leaves the snapshot is reported as a lost selection', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);
    viewer.selectIdentity('addr:aa:bb:cc:dd:ee:01');

    state.value = snapshot([]);

    expect(viewer.selected.value).toBeNull();
    expect(viewer.selectionLost.value).toBe(true);
  });

  test('focus is not offered without a selection', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = useObservationViewer(() => state.value);

    expect(viewer.availableViews.value).not.toContain(ObservationViews.Focus);
  });

  test('every view has a label', () => {
    for (const view of Object.values(ObservationViews)) {
      expect(OBSERVATION_VIEW_LABELS[view]).toBeTruthy();
    }
  });
});
