<script setup lang="ts">
  /*
   * Home (spec 007 US-02, US-06 / FR-09). Areas are shown in their saved order. In "Edit
   * layout" mode areas can be dragged by their handle or moved with the arrow buttons
   * (keyboard and touch friendly); each change saves the new positions to the hub.
   */
  import { computed, ref, watch } from 'vue';
  import draggable from 'vuedraggable';
  import { DashboardGroup, DashboardGroups } from '@/types/settings.type';
  import { RouteName } from '@/types/router';
  import { store } from '@/store';
  import { useDashboardGroups } from '@/composables/useDashboardGroups';
  import { useHub } from '@/composables/useHub';
  import { confirm } from '@/composables/useConfirm';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import AreaCard from './AreaCard.vue';
  import AreaNameDialog from './AreaNameDialog.vue';
  import AddTilesDialog from './AddTilesDialog.vue';

  const areas = useDashboardGroups();
  const { initialized } = useHub();

  const editing = ref(false);
  /** Local copy the drag list mutates; resynced from the store whenever it changes. */
  const list = ref<DashboardGroup[]>([]);
  watch(areas.groups, (groups) => (list.value = [...groups]), { immediate: true });

  const tileCount = computed(() => list.value.reduce((n, g) => n + Object.values(g.deviceGroup ?? {}).reduce((m, d) => m + d.exposes.length, 0), 0));

  // dialogs
  const nameDialog = ref<{ open: boolean; current?: string }>({ open: false });
  const tilesFor = ref<DashboardGroup | null>(null);

  function saveName(name: string) {
    const current = nameDialog.value.current;
    if (current) areas.rename(current, name);
    else areas.create(name);
    nameDialog.value = { open: false };
  }

  function onDragEnd() {
    areas.applyOrder(list.value);
  }

  async function removeArea(group: DashboardGroup) {
    const ok = await confirm({
      title: `Delete ${group.name}?`,
      message: 'The area and its tiles are removed from Home. Devices are not affected.',
      confirmLabel: 'Delete area',
      danger: true,
    });
    if (ok) areas.remove(group.name);
  }

  function addTiles(deviceId: string, exposes: string[]) {
    if (tilesFor.value) areas.addEntities(tilesFor.value, deviceId, exposes);
    tilesFor.value = null;
  }

  function exportAreas() {
    const url = URL.createObjectURL(new Blob([areas.exportJson()], { type: 'application/json' }));
    const link = document.createElement('a');
    link.href = url;
    link.download = 'dashboard-groups.json';
    link.click();
    URL.revokeObjectURL(url);
  }

  const fileInput = ref<HTMLInputElement | null>(null);
  async function importAreas(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    let groups: DashboardGroups | undefined;
    try {
      groups = JSON.parse(await file.text())?.dashboardGroups;
    } catch {
      groups = undefined;
    }
    if (!groups || typeof groups !== 'object' || Array.isArray(groups)) {
      store.dispatch('alerts/showError', 'That file is not a NodeHerder areas export.');
      return;
    }
    const ok = await confirm({
      title: 'Import areas?',
      message: `Import ${Object.keys(groups).length} area(s) from ${file.name}. Areas with the same name are replaced.`,
      confirmLabel: 'Import',
    });
    if (ok) areas.importGroups(groups);
  }
</script>

<template>
  <div class="nh-home">
    <UiPageHeader title="Home" :subtitle="`${list.length} areas · ${tileCount} tiles`">
      <template #actions>
        <template v-if="editing">
          <UiButton icon="add" @click="nameDialog = { open: true }">New area</UiButton>
          <UiButton icon="export" :disabled="!list.length" @click="exportAreas">Export</UiButton>
          <UiButton icon="import" @click="fileInput?.click()">Import</UiButton>
          <UiButton variant="primary" icon="check" @click="editing = false">Done</UiButton>
        </template>
        <template v-else>
          <UiButton icon="panel" :to="{ name: RouteName.Panel }">Panel mode</UiButton>
          <UiButton icon="edit" @click="editing = true">Edit layout</UiButton>
        </template>
        <input ref="fileInput" type="file" accept=".json,application/json" class="sr-only" tabindex="-1" aria-hidden="true" @change="importAreas" />
      </template>
    </UiPageHeader>

    <p v-if="editing" class="nh-alert is-info nh-home-hint" role="status">
      Drag an area by its handle, or use the arrows, to change the order. Changes save as you go.
    </p>

    <div v-if="!initialized && !list.length" class="nh-areas" aria-busy="true">
      <div v-for="n in 3" :key="n" class="nh-card nh-skel-card"><span class="nh-skel" /><span class="nh-skel" /></div>
    </div>
    <UiEmpty v-else-if="!list.length" icon="home" title="No areas yet" text="Areas group the values you check most, like a room.">
      <UiButton variant="primary" icon="add" @click="nameDialog = { open: true }">Create your first area</UiButton>
    </UiEmpty>
    <draggable
      v-else
      v-model="list"
      item-key="name"
      class="nh-areas"
      handle=".nh-area-handle"
      :disabled="!editing"
      :animation="160"
      ghost-class="is-ghost"
      @end="onDragEnd">
      <template #item="{ element, index }">
        <AreaCard
          :group="element"
          :editing="editing"
          :index="index"
          :count="list.length"
          @move="areas.move(index, $event)"
          @rename="nameDialog = { open: true, current: element.name }"
          @remove="removeArea(element)"
          @add-tiles="tilesFor = element"
          @remove-tile="(deviceId: string, expose: string) => areas.removeEntity(element, deviceId, expose)" />
      </template>
    </draggable>

    <AreaNameDialog :open="nameDialog.open" :current="nameDialog.current" :validate="areas.validateName" @close="nameDialog = { open: false }" @save="saveName" />
    <AddTilesDialog :open="!!tilesFor" :group="tilesFor" @close="tilesFor = null" @add="addTiles" />
  </div>
</template>

<style scoped>
  .nh-areas {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, var(--nh-area-w)), 1fr));
    gap: calc(var(--nh-space) * 1.5) calc(var(--nh-space) * 1.4);
    align-items: start;
  }
  .nh-home-hint {
    margin-bottom: calc(var(--nh-space) * 1.2);
  }
  .nh-skel-card {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 1rem;
  }
  .nh-areas :deep(.is-ghost) {
    opacity: 0.4;
  }
</style>
