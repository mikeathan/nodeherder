import { spatialLabelMismatch, spatialOptions, spatialTargetStatuses, spatialTargets, stableSpatialInventory, validateSpatialMappings } from '@/domain/spatial';
import type { Expose } from '@/types/device';

const expose = (data: unknown, unit: string) => ({ data, unit }) as unknown as Expose;
const mapping = [{ x: 'x', y: 'y', active: 'active' }];

test('keeps selector inventory stable when only live readings change', () => {
  const source = {
    name: '00000000-00000001',
    description: 'Target 2 X',
    unit: 'mm',
    data: 100,
    type: 'numeric',
    category: 'measurement',
    access_mode: 'read',
    attributes: { label: 'Target 2 X' },
    values: {},
  } as unknown as Expose;
  const first = stableSpatialInventory(undefined, [{ ...source, attributes: { ...source.attributes } }]);
  source.data = 200;
  expect(stableSpatialInventory(first, [source])).toBe(first);
  source.attributes = { label: 'Target 2 Y' } as never;
  expect(stableSpatialInventory(first, [source])).not.toBe(first);
});

test('offers matching numbered fields or unrecognized labels for other devices', () => {
  const labelled = (label: string) => ({ name: label, attributes: { label } }) as unknown as Expose;
  const fields = ['Target 1 X', 'Target 1 Y', 'Target 2 X', 'Position'].map(labelled);
  expect(spatialOptions(fields, 1, 'x').map(e => e.name)).toEqual(['Target 1 X']);
  expect(spatialOptions(fields, 1, 'y').map(e => e.name)).toEqual(['Target 1 Y']);
  expect(spatialOptions(fields, 2, 'x').map(e => e.name)).toEqual(['Target 2 X']);
  expect(spatialOptions(fields, 3, 'x').map(e => e.name)).toEqual(['Position']);
  expect(spatialOptions(fields.slice(0, 3), 3, 'x')).toEqual([]);
});

test('explains inactive, stale and unmapped target rows without hiding mapping controls', () => {
  const values = {
    x: expose(100, 'mm'), y: expose(200, 'mm'), active: expose(false, ''),
    x2: expose(300, 'mm'), y2: expose(400, 'mm'), active2: expose(true, ''),
  };
  const rows = [{ x: 'x', y: 'y', active: 'active' }, { x: 'x2', y: 'y2', active: 'active2' }, { x: '', y: '', active: '' }];
  expect(spatialTargetStatuses(values, rows, new Map([['x2', 9000], ['y2', 9500]]), 10000, true)).toEqual([
    { id: 1, state: 'not-active', label: 'Not active' },
    { id: 2, state: 'active', label: 'Active' },
    { id: 3, state: 'incomplete', label: 'Not mapped' },
  ]);
  expect(spatialTargetStatuses(values, rows, new Map(), 10000, true)[1]).toEqual({
    id: 2,
    state: 'stale',
    label: 'Waiting for fresh X/Y',
  });
});

test('requires fresh independent coordinates and an active signal', () => {
  const values = { x: expose(-100, 'cm'), y: expose(2, 'm'), active: expose(true, '') };
  expect(spatialTargets(values, mapping, new Map([['x', 9000], ['y', 9500]]), 10000, true)).toEqual([{ id: 1, x: -1000, y: 2000, distance: Math.hypot(1000, 2000) }]);
  expect(spatialTargets(values, mapping, new Map([['x', 4000], ['y', 9500]]), 10000, true)).toEqual([]);
  expect(spatialTargets(values, mapping, new Map([['x', 9000], ['y', 9500]]), 10000, false)).toEqual([]);
  values.active.data = false;
  expect(spatialTargets(values, mapping, new Map([['x', 9000], ['y', 9500]]), 10000, true)).toEqual([]);
});

test('detects clearly mismatched target labels without assuming unknown labels', () => {
  const labelled = (label: string) => ({ name: label, attributes: { label } }) as unknown as Expose;
  const values = { x: labelled('Target 3 X'), y: labelled('Target 3 Y'), active: labelled('Target 1 Active'), other: labelled('Position') };
  expect(spatialLabelMismatch({ x: 'x', y: 'y', active: 'active' }, values)).toBe('target');
  expect(spatialLabelMismatch({ x: 'y', y: 'x', active: '' }, values)).toBe('axis');
  values.active = labelled('Target 3 Active');
  expect(spatialLabelMismatch({ x: 'x', y: 'y', active: 'active' }, values)).toBeNull();
  expect(spatialLabelMismatch({ x: 'x', y: 'y', active: 'active' }, values, 1)).toBe('row');
  expect(spatialLabelMismatch({ x: 'x', y: 'y', active: 'active' }, values, 3)).toBeNull();
  expect(spatialLabelMismatch({ x: 'x', y: 'y', active: 'other' }, values)).toBeNull();
});

test('does not display a target assembled from different labelled targets', () => {
  const labelled = (label: string, data: unknown, unit: string) => ({ name: label, attributes: { label }, data, unit }) as unknown as Expose;
  const values = { x: labelled('Target 3 X', 100, 'mm'), y: labelled('Target 3 Y', 200, 'mm'), active: labelled('Target 1 Active', true, '') };
  expect(spatialTargets(values, mapping, new Map([['x', 9000], ['y', 9000]]), 10000, true)).toEqual([]);
});

 test('mapping validation distinguishes optional, incomplete and unreadable coordinates', () => {
 const fields = {
 x: { name: 'x', type: 'numeric', unit: 'mm', access_mode: 'read' },
 y: { name: 'y', type: 'numeric', unit: 'm', access_mode: 'read' },
 active: { name: 'active', type: 'binary', access_mode: 'read' },
 } as unknown as Record<string, Expose>;
 expect(validateSpatialMappings([{ x: '', y: '', active: '' }], fields)).toBe('');
 expect(validateSpatialMappings(mapping, fields)).toBe('');
 expect(validateSpatialMappings([{ x: 'x', y: '', active: '' }], fields)).toContain('choose an X, Y and Active');
 expect(validateSpatialMappings([{ x: 'x', y: 'x', active: 'active' }], fields)).toContain('different entities');
 fields.y.access_mode = 'write' as never;
 expect(validateSpatialMappings(mapping, fields)).toContain('no longer available');
 });
