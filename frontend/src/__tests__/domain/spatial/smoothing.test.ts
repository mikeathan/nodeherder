import {
  SPATIAL_DEFAULT_SMOOTHING,
  SPATIAL_SMOOTHING_LABELS,
  SPATIAL_SMOOTHING_OPTIONS,
  SPATIAL_SMOOTHING_PRESETS,
} from '@/domain/spatial/smoothing';

test('smoothing levels are ordered, labelled and offered', () => {
  expect(SPATIAL_SMOOTHING_PRESETS.low.minCutoff).toBeGreaterThan(SPATIAL_SMOOTHING_PRESETS.medium.minCutoff);
  expect(SPATIAL_SMOOTHING_PRESETS.medium.minCutoff).toBeGreaterThan(SPATIAL_SMOOTHING_PRESETS.high.minCutoff);
  expect(SPATIAL_SMOOTHING_PRESETS.medium.beta).toBeGreaterThan(SPATIAL_SMOOTHING_PRESETS.high.beta);
  expect(SPATIAL_SMOOTHING_OPTIONS.map((option) => option.value)).toEqual(['off', 'low', 'medium', 'high']);
  for (const option of SPATIAL_SMOOTHING_OPTIONS) {
    expect(option.label).toBe(SPATIAL_SMOOTHING_LABELS[option.value]);
    expect(option.label.length).toBeGreaterThan(0);
  }
  expect(SPATIAL_DEFAULT_SMOOTHING).toBe('medium');
});
