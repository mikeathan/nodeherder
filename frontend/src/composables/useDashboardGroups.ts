/*
 * Home areas (dashboard groups) for screens (spec 007 US-02, US-06 / FR-09). Wraps
 * domain/dashboard.ts and the hub store actions; every change is sent to the hub with the
 * existing WebSocket commands.
 */
import { computed } from 'vue';
import { store } from '@/store';
import { DashboardGroup, DashboardGroups } from '@/types/settings.type';
import { moveItem, nextOrder, reorderedGroups, sortGroups, validateGroupName, withEntities, withoutEntity } from '@/domain/dashboard';

const allGroups = () => store.getters['hub/dashboardGroups']() as DashboardGroups;

export function useDashboardGroups() {
  const groups = computed(() => sortGroups(allGroups() ?? {}));

  const save = (group: DashboardGroup) => store.dispatch('hub/saveDashboardGroup', group);

  /** Saves the new positions after a move; only groups whose position changed are sent (AC-19). */
  function applyOrder(ordered: readonly DashboardGroup[]) {
    reorderedGroups(ordered).forEach(save);
  }

  return {
    groups,
    validateName: (name: string, current?: string) => validateGroupName(name, allGroups(), current),
    move(from: number, to: number) {
      if (from === to) return;
      applyOrder(moveItem(groups.value, from, to));
    },
    applyOrder,
    create(name: string) {
      const trimmed = name.trim();
      save({ name: trimmed, deviceGroup: {}, order: nextOrder(allGroups()) });
    },
    rename(oldName: string, newName: string) {
      const trimmed = newName.trim();
      if (trimmed && trimmed !== oldName) store.dispatch('hub/renameDashboardGroup', { oldName, newName: trimmed });
    },
    remove(name: string) {
      store.dispatch('hub/deleteDashboardGroup', name);
    },
    addEntities(group: DashboardGroup, deviceId: string, exposes: string[]) {
      save(withEntities(group, deviceId, exposes));
    },
    removeEntity(group: DashboardGroup, deviceId: string, expose: string) {
      save(withoutEntity(group, deviceId, expose));
    },
    exportJson(): string {
      return JSON.stringify({ dashboardGroups: allGroups() }, null, 2);
    },
    importGroups(groups: DashboardGroups) {
      store.dispatch('hub/importDashboardGroups', groups);
    },
  };
}
