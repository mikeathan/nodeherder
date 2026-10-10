import { describe, expect, test } from '@jest/globals';
import { DEFAULT_FILTERS, filterDevices, needsAttention, sortDevices, summarizeNetwork } from '@/domain/network';
import { device, expose, hubDevices } from '../helpers/devices';

const never = () => false;

describe('network summary', () => {
  test('counts the fixture network', () => {
    const s = summarizeNetwork(hubDevices());
    expect(s.total).toBe(13);
    expect(s.online + s.offline).toBe(13);
    expect(s.zigbeeRouters + s.zigbeeEndDevices).toBe(13);
    expect(s.otherProtocols).toBe(0);
  });
  test('weak link and low battery', () => {
    const s = summarizeNetwork([device('a', {}, [expose('linkquality', { data: 10 })]), device('b', { power_source: 'battery', connection_type: 'http' }, [expose('battery', { data: 5 })])]);
    expect(s).toMatchObject({ weakLink: 1, lowBattery: 1, zigbeeRouters: 1, otherProtocols: 1 });
  });
});

describe('needs attention', () => {
  test('orders offline, low battery, weak link, disabled, then by name', () => {
    const items = needsAttention(
      [device('weak', {}, [expose('linkquality', { data: 3 })]), device('dis'), device('off', { availability: 'offline' as never }), device('bat', {}, [expose('battery', { data: 2 })]), device('fine')],
      (d) => d.id === 'dis'
    );
    expect(items.map((i) => [i.device.id, i.reason])).toEqual([['off', 'offline'], ['bat', 'low-battery'], ['weak', 'weak-link'], ['dis', 'disabled']]);
  });
});

describe('filter and sort', () => {
  const list = [
    device('z1', { friendly_name: 'Garage weather', connection_type: 'http', last_seen: '2026-10-10T10:05:00Z' }, [expose('rssi', { data: -60 })]),
    device('z2', { friendly_name: 'attic light', power_source: 'battery', availability: 'offline' as never, last_seen: '2026-10-10T09:00:00Z' }, [expose('linkquality', { data: 80 }), expose('battery', { data: 50 })]),
    device('z3', { friendly_name: 'Bedroom valve', power_source: 'battery', last_seen: 'not a date' }, [expose('linkquality', { data: 20 }), expose('battery', { data: 10 })]),
  ];
  test('query matches name, id and description; protocol, power and status filters', () => {
    expect(filterDevices(list, { ...DEFAULT_FILTERS, query: 'ATTIC' }, never).map((d) => d.id)).toEqual(['z2']);
    expect(filterDevices(list, { ...DEFAULT_FILTERS, protocol: 'http' }, never).map((d) => d.id)).toEqual(['z1']);
    expect(filterDevices(list, { ...DEFAULT_FILTERS, power: 'mains' }, never).map((d) => d.id)).toEqual(['z1']);
    expect(filterDevices(list, { ...DEFAULT_FILTERS, status: 'offline' }, never).map((d) => d.id)).toEqual(['z2']);
    expect(filterDevices(list, { ...DEFAULT_FILTERS, status: 'disabled' }, (d) => d.id === 'z3').map((d) => d.id)).toEqual(['z3']);
    expect(filterDevices(list, { ...DEFAULT_FILTERS, status: 'online' }, (d) => d.id === 'z3').map((d) => d.id)).toEqual(['z1']);
  });
  test('sorts by name case-insensitively, signal/battery with missing last both directions, recent first', () => {
    expect(sortDevices(list, 'name').map((d) => d.id)).toEqual(['z2', 'z3', 'z1']);
    expect(sortDevices(list, 'name', false).map((d) => d.id)).toEqual(['z1', 'z3', 'z2']);
    expect(sortDevices(list, 'signal').map((d) => d.id)).toEqual(['z3', 'z2', 'z1']);
    expect(sortDevices(list, 'signal', false).map((d) => d.id)).toEqual(['z2', 'z3', 'z1']);
    expect(sortDevices(list, 'battery').map((d) => d.id)).toEqual(['z3', 'z2', 'z1']);
    expect(sortDevices(list, 'lastSeen').map((d) => d.id)).toEqual(['z1', 'z2', 'z3']);
  });
});
