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

/**
 * Every viewer created here starts a one-second interval that only a component unmount clears.
 * Nothing in this file is a component, so each instance leaked a handle and Jest could not exit —
 * which is how the leak surfaced. The suite now closes them itself.
 */
const stopClocks: { stop(): void }[] = [];
function createViewer(next: () => ObservationSnapshot | null) {
  const instance = useObservationViewer(next);
  stopClocks.push(instance);
  return instance;
}
afterEach(() => {
  for (const instance of stopClocks.splice(0)) instance.stop();
});

describe('useObservationViewer', () => {
  test('renders the live snapshot when not paused', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);

    expect(viewer.visible.value).toHaveLength(1);
    expect(viewer.paused.value).toBe(false);
  });

  test('groups facets declared by sources, not from a hardcoded list', () => {
    const state = ref(snapshot([], ['kind', 'addressType']));
    const viewer = createViewer(() => state.value);

    expect(viewer.facets.value.sort()).toEqual(['addressType', 'kind']);
  });

  test('pausing freezes the rendered identities', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);

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
    const viewer = createViewer(() => state.value);

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
    const viewer = createViewer(() => state.value);

    viewer.togglePause();
    viewer.applyDecisionToFrozen('addr:aa:bb:cc:dd:ee:01', ObservationDecisions.Known);

    expect(viewer.visible.value[0]?.decision).toBe(ObservationDecisions.Known);
  });

  test('a decision outside the frozen view changes nothing', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);

    // Not paused: the frozen list does not exist, so this must be a no-op rather than a crash.
    expect(() => viewer.applyDecisionToFrozen('missing', ObservationDecisions.Known)).not.toThrow();
  });

  test('selecting an identity isolates it and switches to focus', () => {
    const state = ref(
      snapshot([identity('addr:aa:bb:cc:dd:ee:01'), identity('addr:aa:bb:cc:dd:ee:02')]),
    );
    const viewer = createViewer(() => state.value);

    viewer.selectIdentity('addr:aa:bb:cc:dd:ee:02');

    expect(viewer.scoped.value).toHaveLength(1);
    expect(viewer.viewMode.value).toBe(ObservationViews.Focus);
    expect(viewer.availableViews.value).toContain(ObservationViews.Focus);
  });

  test('clearing a selection leaves focus', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);
    viewer.selectIdentity('addr:aa:bb:cc:dd:ee:01');

    viewer.clearSelection();

    expect(viewer.selected.value).toBeNull();
    expect(viewer.viewMode.value).toBe(ObservationViews.Radar);
  });

  test('an identity that leaves the snapshot is reported as a lost selection', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);
    viewer.selectIdentity('addr:aa:bb:cc:dd:ee:01');

    state.value = snapshot([]);

    expect(viewer.selected.value).toBeNull();
    expect(viewer.selectionLost.value).toBe(true);
  });

  test('focus is not offered without a selection', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);

    expect(viewer.availableViews.value).not.toContain(ObservationViews.Focus);
  });

  test('every view has a label', () => {
    for (const view of Object.values(ObservationViews)) {
      expect(OBSERVATION_VIEW_LABELS[view]).toBeTruthy();
    }
  });
});

/**
 * The default filter is the point of classification.
 *
 * The operator's complaint was "a big list of mostly mac address". Installed equipment and
 * neighbours are the bulk of that list, so the default view hides them — but the count is reported
 * and the toggle exists, because a device hidden with no way to see it is a device that looks lost.
 */
describe('demoted classes', () => {
  function classified(key: string, showByDefault: boolean) {
    return identity(key, {
      classification: {
        class: showByDefault ? 'personal' : 'infrastructure',
        showByDefault,
        confidence: 0.7,
        hoursObserved: 30,
        advisory: true,
      },
    });
  }

  test('hides demoted identities by default', () => {
    const state = ref(
      snapshot([
        classified('addr:aa:bb:cc:dd:ee:01', true),
        classified('addr:aa:bb:cc:dd:ee:02', false),
        classified('addr:aa:bb:cc:dd:ee:03', false),
      ]),
    );
    const viewer = createViewer(() => state.value);

    expect(viewer.visible.value).toHaveLength(1);
    expect(viewer.demotedCount.value).toBe(2);
  });

  test('reveals them when asked', () => {
    const state = ref(
      snapshot([
        classified('addr:aa:bb:cc:dd:ee:01', true),
        classified('addr:aa:bb:cc:dd:ee:02', false),
      ]),
    );
    const viewer = createViewer(() => state.value);

    viewer.showDemoted.value = true;

    expect(viewer.visible.value).toHaveLength(2);
    expect(viewer.demotedCount.value).toBe(0);
  });

  test('shows an unclassified identity rather than hiding it', () => {
    // An older backend sends no classification. Hiding those would blank the page on a mixed
    // deployment, and showing too much is the safe failure.
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:09')]));
    const viewer = createViewer(() => state.value);

    expect(viewer.visible.value).toHaveLength(1);
    expect(viewer.demotedCount.value).toBe(0);
  });

  test('counts only the demoted, not everything filtered out', () => {
    // The footer says how many are hidden by class. A distance filter hiding a device is a
    // different statement, and conflating them would make the number meaningless.
    const state = ref(
      snapshot([
        classified('addr:aa:bb:cc:dd:ee:01', true),
        classified('addr:aa:bb:cc:dd:ee:02', false),
      ]),
    );
    const viewer = createViewer(() => state.value);

    viewer.onlyNew.value = true;

    expect(viewer.demotedCount.value).toBe(1);
  });
});

describe('demoted toggle availability', () => {
  test('hasDemoted is false when nothing is demoted', () => {
    // Before classification has run every device is learning or provisional, so nothing is hidden
    // and the reveal control would toggle between two identical states — which reads as broken.
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);

    expect(viewer.hasDemoted.value).toBe(false);
    expect(viewer.demotedCount.value).toBe(0);
  });

  test('hasDemoted stays true while revealed, so the toggle does not vanish mid-use', () => {
    const state = ref(
      snapshot([
        identity('addr:aa:bb:cc:dd:ee:02', {
          classification: {
            class: 'infrastructure',
            showByDefault: false,
            confidence: 0.8,
            hoursObserved: 30,
            advisory: true,
          },
        }),
      ]),
    );
    const viewer = createViewer(() => state.value);

    expect(viewer.hasDemoted.value).toBe(true);
    viewer.showDemoted.value = true;
    // The count drops to zero because nothing is hidden now, but the control must stay: otherwise
    // the button disappears the moment it is pressed and there is no way back.
    expect(viewer.demotedCount.value).toBe(0);
    expect(viewer.hasDemoted.value).toBe(true);
  });
});
