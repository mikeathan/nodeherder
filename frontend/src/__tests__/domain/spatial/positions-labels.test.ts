import { describe, expect, test } from '@jest/globals';
import {
  countLabel,
  distanceProgressLabel,
  emptyRecordingLabel,
  fallbackMarkLabel,
  hasPlayableSpan,
  idleRecordingLabel,
  movementChapterLabel,
  movementCountLabel,
  movementMarkLabel,
  playbackValueLabel,
  railMetaLabel,
  railTimeLabel,
  rangeHeadlineLabel,
  segmentLabel,
  skippedGapLabel,
  timeLabelOptions,
  transportStatusLabel,
} from '@/domain/spatial/positions-labels';
import { clock, type SessionSegment } from '@/domain/spatial/playback';
import type { SpatialSession } from '@/services/spatial-history.service';

function session(overrides: Partial<SpatialSession>): SpatialSession {
  return { id: 's', from: 0, to: 1, points: 1, slots: [1], capped: false, ...overrides };
}

function segment(overrides: Partial<SessionSegment>): SessionSegment {
  return {
    id: 's',
    from: 0,
    to: 1_000,
    leftPct: 0,
    widthPct: 10,
    points: 1,
    slots: [1],
    capped: false,
    activity: [],
    sampleCount: 0,
    ...overrides,
  };
}

describe('positions labels', () => {
  test('pluralises once for the whole family', () => {
    expect(countLabel(1, 'recording')).toBe('1 recording');
    expect(countLabel(2, 'recording')).toBe('2 recordings');
    expect(countLabel(1, 'recorded position')).toBe('1 recorded position');
    expect(countLabel(0, 'sample')).toBe('0 samples');
  });

  test('treats a single-sample recording as nothing to scrub', () => {
    expect(hasPlayableSpan(session({ from: 1_000, to: 90_000 }))).toBe(true);
    // One sample, or all samples at the same instant: no span worth rendering a scrubber for.
    expect(hasPlayableSpan(session({ from: 1_000, to: 1_000 }))).toBe(false);
    expect(hasPlayableSpan(session({ from: 1_000, to: 1_400 }))).toBe(false);
  });

  test('chooses date and second precision from the window', () => {
    expect(timeLabelOptions(86_400_001).date).toBe(true);
    expect(timeLabelOptions(86_400_000).date).toBe(false);
    expect(timeLabelOptions(3_599_999).seconds).toBe(true);
    expect(timeLabelOptions(3_600_000).seconds).toBe(false);
  });

  test('says an empty-looking recording lists what its summary claimed', () => {
    expect(emptyRecordingLabel(0)).toBe('No recorded positions in this session.');
    expect(emptyRecordingLabel(8)).toBe(
      'This recording lists 8 positions in this range, but none could be loaded.'
    );
    expect(emptyRecordingLabel(1)).toContain('1 position');
  });

  test('labels a rail row as one sample rather than a zero duration', () => {
    const single = railMetaLabel(session({ from: 5, to: 5, points: 1 }));
    expect(single).toBe('1 sample · slot 1');
    const many = railMetaLabel(session({ from: 0, to: 120_000, points: 4, slots: [1, 2] }));
    expect(many).toBe('2 min · slots 1, 2');
    // A window under an hour needs seconds, or four recordings print identical ranges; at an
    // hour and beyond the date/minute precision carries enough.
    expect(railTimeLabel(session({ from: 0, to: 60_000 }), 3_599_999)).toMatch(/\d{2}:\d{2}:\d{2}/);
    expect(railTimeLabel(session({ from: 0, to: 60_000 }), 3_600_000)).not.toMatch(/\d{2}:\d{2}:\d{2}/);
    expect(railTimeLabel(session({ from: 0, to: 60_000 }), 86_400_000)).not.toMatch(/\d{2}:\d{2}:\d{2}/);
  });

  test('summarises movements and the range head', () => {
    expect(movementCountLabel(0)).toBe('No movement recorded in this recording.');
    expect(movementCountLabel(1)).toBe('1 movement in this recording');
    expect(movementCountLabel(8, 4)).toBe('8 movements in this recording, 4 shown');
    expect(rangeHeadlineLabel(1, 0, 60_000)).toContain('1 recording between');
    expect(rangeHeadlineLabel(2, 0, 60_000)).toContain('2 recordings between');
  });

  test('builds the segment accessible name with every fact a pointer user cannot see', () => {
    const label = segmentLabel(
      segment({ from: 0, to: 120_000, points: 4, sampleCount: 3, slots: [1, 2], capped: true }),
      false
    );
    expect(label).toContain('2 min');
    expect(label).toContain('4 points');
    expect(label).toContain('slot 1, 2');
    expect(label).toContain('capped');
    expect(label).toContain('3 recorded positions');
  });

  test('buckets the idle line so a live region does not repeat every tick', () => {
    // Under the first ten-second bucket there is nothing worth announcing.
    expect(idleRecordingLabel(400, 0)).toBe('');
    expect(idleRecordingLabel(1_400, 0)).toBe('');
    // Ten-second buckets: the playhead advances every 200 ms, so the sentence must not change
    // five times a second (nor every second over a long tail).
    expect(idleRecordingLabel(14_000, 0)).toContain('10 s');
    expect(idleRecordingLabel(14_000, 0)).not.toContain('14000');
    expect(idleRecordingLabel(75_000, 0)).toContain('1 min');
  });

  test('ranks the transport states: a crossed gap outranks idle, which outranks context', () => {
    const base = {
      skippedGap: null,
      pendingGap: null,
      session: session({ from: 0, to: 60_000 }),
      idleForMs: 0,
      lastSampleAt: 0,
      awaitingCrossing: false,
      reachedEnd: false,
      movementCount: 3,
      renderedMovements: 3,
    };
    expect(transportStatusLabel({ ...base, skippedGap: { from: 0, to: 9_000 } })).toContain(
      'Skipped'
    );
    expect(
      transportStatusLabel({ ...base, session: session({ from: 5, to: 5 }) })
    ).toContain('Nothing to play or scrub');
    expect(transportStatusLabel({ ...base, idleForMs: 15_000 })).toContain('No movement recorded');
    expect(transportStatusLabel({ ...base, awaitingCrossing: true })).toContain('Loading');
    expect(transportStatusLabel({ ...base, reachedEnd: true })).toContain('End of the filtered');
    expect(
      transportStatusLabel({ ...base, pendingGap: { from: 0, to: 9_000 } })
    ).toContain('Next recording is after');
    expect(transportStatusLabel(base)).toContain('3 movements in this recording');
  });

  test('names a per-movement mark by its time and recording, without inventing a slot', () => {
    const label = movementMarkLabel(
      segment({ from: 1_000_000, to: 1_120_000 }),
      1_060_000,
      false
    );
    expect(label).toContain('Movement at');
    // Local timezone: derive the expected time with the same formatter the label uses.
    expect(label).toContain(clock(1_060_000));
    expect(label).toContain('–');
    // The summary carries timestamps, not per-sample slots.
    expect(label).not.toContain('slot');
  });

  test('labels a skipped gap with its duration and both ends', () => {
    const label = skippedGapLabel({ from: 0, to: 240_000 });
    expect(label).toContain('Skipped 4 min');
    expect(label).toContain('with no recording');
  });

  test('names a scrubber chapter by its time and slot, and the playhead against the end', () => {
    const chapter = movementChapterLabel(1_060_000, 2);
    expect(chapter).toContain('Movement at');
    expect(chapter).toContain(clock(1_060_000));
    expect(chapter).toContain('slot 2');
    expect(playbackValueLabel(1_060_000, 1_120_000)).toBe(
      `${clock(1_060_000)} of ${clock(1_120_000)}`
    );
  });

  test('states distance progress through the shared unit formatter', () => {
    // Millimetres in, human units out: this must not be the component's job.
    expect(distanceProgressLabel(350, 1_200_000)).toBe('0.35 m of 1.2 km');
  });

  test('names the fallback control a recording keeps when it has no measurable marks', () => {
    const label = fallbackMarkLabel(segment({ from: 1_000_000, to: 1_120_000 }), false);
    expect(label).toContain('Play from');
    expect(label).toContain('–');
    // A fallback is a selection target, never an invented position.
    expect(label).not.toContain('Movement at');
  });
});
