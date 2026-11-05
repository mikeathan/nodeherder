import {
  REPLAY_MAX_FRAMES,
  REPLAY_WINDOW_MS,
  appendReplayFrame,
  frameAt,
  replayTracks,
} from '@/domain/spatial/replay';
import type { ReplayFrame } from '@/domain/spatial/replay';

test('recording keeps a bounded window and frame count', () => {
  let frames: ReplayFrame[] = Array.from({ length: 5 }, (_, index) => ({ at: index * 1000, samples: [] }));
  frames = appendReplayFrame(frames, { at: 60_000, samples: [] });
  expect(frames.at(-1)?.at).toBe(60_000);
  expect(frames.some((frame) => frame.at <= 60_000 - REPLAY_WINDOW_MS)).toBe(false);
  const many = appendReplayFrame(
    Array.from({ length: REPLAY_MAX_FRAMES }, (_, index) => ({ at: index, samples: [] })),
    { at: REPLAY_MAX_FRAMES, samples: [] }
  );
  expect(many).toHaveLength(REPLAY_MAX_FRAMES);
  expect(many.at(-1)?.at).toBe(REPLAY_MAX_FRAMES);
});

test('seeking clamps to the recorded range', () => {
  const frames = [
    { at: 1000, samples: [] },
    { at: 2000, samples: [] },
    { at: 3000, samples: [] },
  ];
  expect(frameAt(frames, 0)?.at).toBe(1000);
  expect(frameAt(frames, 1500)?.at).toBe(1000);
  expect(frameAt(frames, 2500)?.at).toBe(2000);
  expect(frameAt(frames, 9000)?.at).toBe(3000);
  expect(frameAt([], 1000)).toBeUndefined();
});

test('replay frames become labelled slot tracks without claiming identity', () => {
  const tracks = replayTracks({
    at: 5000,
    samples: [
      { slot: 1, x: 100, y: 2000 },
      { slot: 3, x: -400, y: 1500 },
    ],
  });
  expect(tracks).toHaveLength(2);
  expect(tracks[0]).toMatchObject({ slot: 1, label: 'Slot 1', rawX: 100, rawY: 2000, at: 5000, state: 'live' });
  expect(tracks[1].colorIndex).toBe(2);
});
