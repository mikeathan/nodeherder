import { describe, expect, test } from '@jest/globals';
import { groupEntities, moveItem, nextOrder, reorderedGroups, sortGroups, validateGroupName, withEntities, withoutEntity } from '@/domain/dashboard';
import { DashboardGroup, DashboardGroups } from '@/types/settings.type';
import { hubConfig } from '../helpers/devices';

const g = (name: string, order?: number): DashboardGroup => ({ name, deviceGroup: {}, ...(order === undefined ? {} : { order }) });

describe('group ordering (FR-09)', () => {
  test('ordered groups first by order, then unordered by name (AC-21)', () => {
    const groups: DashboardGroups = { b: g('b'), a: g('a'), z: g('z', 0), y: g('y', 1) };
    expect(sortGroups(groups).map((x) => x.name)).toEqual(['z', 'y', 'a', 'b']);
  });
  test('fixture groups without order sort by name case-insensitively', () => {
    expect(sortGroups(hubConfig().hub.dashboardGroups).map((x: DashboardGroup) => x.name)).toEqual(['attic room', 'Kitchen', 'living room group']);
  });
  test('moveItem moves and clamps, never mutates', () => {
    const list = ['a', 'b', 'c'];
    expect(moveItem(list, 0, 2)).toEqual(['b', 'c', 'a']);
    expect(moveItem(list, 2, -5)).toEqual(['c', 'a', 'b']);
    expect(moveItem(list, 9, 0)).toEqual(['a', 'b', 'c']);
    expect(list).toEqual(['a', 'b', 'c']);
  });
  test('reorderedGroups returns only groups whose order changed (AC-19)', () => {
    const changed = reorderedGroups([g('x', 0), g('y', 2), g('z')]);
    expect(changed).toEqual([{ ...g('y'), order: 1 }, { ...g('z'), order: 2 }]);
    expect(reorderedGroups([g('x', 0), g('y', 1)])).toEqual([]);
  });
  test('first reorder of unordered data assigns 0..n-1', () => {
    const ordered = moveItem(sortGroups({ a: g('a'), b: g('b'), c: g('c') }), 2, 0);
    expect(reorderedGroups(ordered).map((x) => [x.name, x.order])).toEqual([['c', 0], ['a', 1], ['b', 2]]);
  });
  test('nextOrder', () => {
    expect(nextOrder({})).toBe(0);
    expect(nextOrder({ a: g('a', 3), b: g('b') })).toBe(4);
  });
});

describe('group content', () => {
  const group: DashboardGroup = { name: 'x', deviceGroup: { d1: { deviceId: 'd1', exposes: ['temperature', 'humidity'] }, d2: { deviceId: 'd2', exposes: ['state'] } } };
  test('entities flatten in stored order', () => {
    expect(groupEntities(group)).toEqual([{ deviceId: 'd1', expose: 'temperature' }, { deviceId: 'd1', expose: 'humidity' }, { deviceId: 'd2', expose: 'state' }]);
  });
  test('remove an entity, dropping empty device groups; add without duplicates', () => {
    expect(withoutEntity(group, 'd2', 'state').deviceGroup.d2).toBeUndefined();
    expect(withoutEntity(group, 'd1', 'humidity').deviceGroup.d1.exposes).toEqual(['temperature']);
    expect(withEntities(group, 'd1', ['humidity', 'battery']).deviceGroup.d1.exposes).toEqual(['temperature', 'humidity', 'battery']);
    expect(group.deviceGroup.d1.exposes).toEqual(['temperature', 'humidity']);
  });
  test('name validation', () => {
    const groups = { Kitchen: g('Kitchen') };
    expect(validateGroupName('  ', groups)).toMatch(/Enter/);
    expect(validateGroupName('kitchen', groups)).toMatch(/exists/);
    expect(validateGroupName('Kitchen', groups, 'Kitchen')).toBeNull();
    expect(validateGroupName('Garage', groups)).toBeNull();
    expect(validateGroupName('x'.repeat(61), groups)).toMatch(/60/);
  });
});
