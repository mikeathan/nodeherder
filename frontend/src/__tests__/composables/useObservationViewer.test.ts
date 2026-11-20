import { describe, expect, test } from '@jest/globals';
import { ref } from 'vue';
import { useObservationViewer, OBSERVATION_VIEW_LABELS } from '@/composables/useObservationViewer';
import type { ObservationSnapshot } from '@/types/observations.type';
import {
  ObservationClasses,
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
    // List is the default surface, so that is where leaving a detail returns to.
    expect(viewer.viewMode.value).toBe(ObservationViews.List);
  });

  test('clearing returns to the view the operator selected from', () => {
    const state = ref(snapshot([identity('addr:aa:bb:cc:dd:ee:01')]));
    const viewer = createViewer(() => state.value);
    viewer.viewMode.value = ObservationViews.Radar;

    viewer.selectIdentity('addr:aa:bb:cc:dd:ee:01');
    viewer.clearSelection();

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

describe('class presets', () => {
  function classified(key: string, classValue: string, showByDefault = true) {
    return identity(key, {
      classification: {
        class: classValue,
        showByDefault,
        confidence: 0.7,
        hoursObserved: 30,
        advisory: true,
      },
    });
  }

  test('offers a preset per class, with the count it would show', () => {
    const state = ref(
      snapshot([
        classified('addr:01', ObservationClasses.Personal),
        classified('addr:02', ObservationClasses.Personal),
        classified('addr:03', ObservationClasses.Learning),
      ]),
    );
    const viewer = createViewer(() => state.value);

    const nonEmpty = viewer.classOptions.value
      .filter((option) => option.count > 0)
      .map((option) => [option.id, option.count]);
    expect(nonEmpty).toEqual([
      ['yours', 2],
      ['learning', 1],
    ]);
  });

  test('keeps every class preset present in fixed order, including zero counts', () => {
    // A chip that disappears at zero, or grows a digit, moves the row while it is being read.
    const state = ref(snapshot([classified('addr:01', ObservationClasses.Personal)]));
    const viewer = createViewer(() => state.value);

    expect(viewer.classOptions.value.map((option) => option.id)).toEqual([
      'yours',
      'someone',
      'installed',
      'learning',
      'outside',
      'passing',
    ]);
    expect(viewer.classOptions.value.filter((option) => option.count === 0)).toHaveLength(5);
  });

  test('selecting a preset narrows the visible set to that group', () => {
    const state = ref(
      snapshot([
        classified('addr:01', ObservationClasses.Personal),
        classified('addr:02', ObservationClasses.Learning),
      ]),
    );
    const viewer = createViewer(() => state.value);

    viewer.classFilter.value = 'yours';

    expect(viewer.visible.value.map((entry) => entry.key)).toEqual(['addr:01']);
  });

  test('"any" restores every identity the other filters allow', () => {
    const state = ref(
      snapshot([
        classified('addr:01', ObservationClasses.Personal),
        classified('addr:02', ObservationClasses.Learning),
      ]),
    );
    const viewer = createViewer(() => state.value);
    viewer.classFilter.value = 'yours';

    viewer.classFilter.value = 'any';

    expect(viewer.visible.value).toHaveLength(2);
  });

  test('an explicit demoted preset reveals its identities without the reveal toggle', () => {
    // Selecting "Installed" is an explicit request. Filtering it to an empty list because a
    // different control is off would look like a broken preset.
    const state = ref(
      snapshot([
        classified('addr:01', ObservationClasses.Infrastructure, false),
        classified('addr:02', ObservationClasses.Personal, true),
      ]),
    );
    const viewer = createViewer(() => state.value);

    expect(viewer.visible.value.map((entry) => entry.key)).toEqual(['addr:02']);

    viewer.classFilter.value = 'installed';

    expect(viewer.visible.value.map((entry) => entry.key)).toEqual(['addr:01']);
    // Nothing is hidden by the default rule while the preset is active, so the footer stays quiet.
    expect(viewer.demotedCount.value).toBe(0);
    // And the now-redundant reveal control is not offered.
    expect(viewer.hasDemoted.value).toBe(false);
  });

  test('counts respect the other filters, so a number is what a click yields', () => {
    const state = ref(
      snapshot([
        classified('addr:01', ObservationClasses.Personal),
        classified('addr:02', ObservationClasses.Personal),
        classified('addr:03', ObservationClasses.Learning),
      ]),
    );
    const viewer = createViewer(() => state.value);

    viewer.search.value = 'addr:02';

    const yours = viewer.classOptions.value.find((option) => option.id === 'yours');
    const learning = viewer.classOptions.value.find((option) => option.id === 'learning');
    expect(yours?.count).toBe(1);
    // The learning group has nothing to show under the query; it stays present so the row does
    // not reflow, but its count is zero and the component disables it.
    expect(learning?.count).toBe(0);
  });
});

describe('search', () => {
  test('filters by key and title, and clearing restores the view', () => {
    const state = ref(
      snapshot([
        identity('addr:aa:bb:cc:dd:ee:01', { label: 'Kitchen lamp' }),
        identity('addr:aa:bb:cc:dd:ee:02', { label: 'Hallway plug' }),
      ]),
    );
    const viewer = createViewer(() => state.value);

    viewer.search.value = 'kitchen';
    expect(viewer.visible.value.map((entry) => entry.key)).toEqual(['addr:aa:bb:cc:dd:ee:01']);

    viewer.search.value = '';
    expect(viewer.visible.value).toHaveLength(2);
  });

  test('composes with the class preset and distance band', () => {
    const state = ref(
      snapshot([
        identity('addr:aa:bb:cc:dd:ee:01', {
          label: 'Phone',
          rssiSmoothed: -60,
          classification: {
            class: ObservationClasses.Personal,
            showByDefault: true,
            confidence: 0.8,
            hoursObserved: 30,
            advisory: true,
          },
        }),
        identity('addr:aa:bb:cc:dd:ee:02', {
          label: 'Phone charger',
          rssiSmoothed: -90,
          classification: {
            class: ObservationClasses.Personal,
            showByDefault: true,
            confidence: 0.8,
            hoursObserved: 30,
            advisory: true,
          },
        }),
      ]),
    );
    const viewer = createViewer(() => state.value);

    viewer.search.value = 'phone';
    viewer.classFilter.value = 'yours';
    viewer.distanceRange.value = '2to5';

    expect(viewer.visible.value).toHaveLength(0);
  });
});

describe('stable ordering', () => {
  test('a volatile-only snapshot update keeps the rendered order', () => {
    const state = ref(
      snapshot([
        identity('addr:aa:bb:cc:dd:ee:02', { label: 'Beta', lastSeen: '2026-11-11T10:00:00Z' }),
        identity('addr:aa:bb:cc:dd:ee:01', { label: 'Alpha', lastSeen: '2026-11-11T09:59:00Z' }),
      ]),
    );
    const viewer = createViewer(() => state.value);
    const before = viewer.visible.value.map((entry) => entry.key);

    state.value = snapshot([
      identity('addr:aa:bb:cc:dd:ee:01', { label: 'Alpha', lastSeen: '2026-11-11T10:00:05Z' }),
      identity('addr:aa:bb:cc:dd:ee:02', { label: 'Beta', lastSeen: '2026-11-11T10:00:04Z' }),
    ]);

    expect(viewer.visible.value.map((entry) => entry.key)).toEqual(before);
  });
});
