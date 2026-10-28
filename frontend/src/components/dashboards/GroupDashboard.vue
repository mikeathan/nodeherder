<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import draggable from 'vuedraggable';
  import type { HomeBlock } from '@/types/home.type';
  import { RouteName } from '@/types/router';
  import { useHomeLayout } from '@/composables/useHomeLayout';
  import { useHomeEditor } from '@/composables/useHomeEditor';
  import { useHomeLayoutFiles } from '@/composables/useHomeLayoutFiles';
  import { useHomeGrid } from '@/composables/useHomeGrid';
  import { useHub } from '@/composables/useHub';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import AreaCard from './AreaCard.vue';
  import AreaNameDialog from './AreaNameDialog.vue';
  import AddTilesDialog from './AddTilesDialog.vue';
  import MoveCardDialog from './MoveCardDialog.vue';
  import HomeStandaloneCard from './HomeStandaloneCard.vue';

  const editing = ref(false);
  const home = useHomeLayout({ pauseRefresh: editing });
  const editor = useHomeEditor(home, editing);
  const files = useHomeLayoutFiles(home);
  const { layout, sources, busy, loading, writable, error, sourceError } = home;
  const { areaDialogOpen, areaToRename, addDialog, addDestination, cardToMove } = editor;
  const { initialized } = useHub();
  const fileInput = ref<HTMLInputElement | null>(null);
  const dragging = ref(false);
  // Only this interaction copy is mutable by the drag library. Accepted membership stays in Vuex.
  const blocks = ref<HomeBlock[]>([]);
  watch([layout, busy], synchronizeBlockOrder, { immediate: true });
  const gridChanges = computed(() => [blocks.value, editing.value, dragging.value]);
  const gridRoot = useHomeGrid(gridChanges);
  const tileCount = computed(() => blocks.value.reduce((total, block) => total + block.cards.length, 0));
  const areaCount = computed(() => blocks.value.filter((block) => block.kind === 'group').length);
  const feedback = computed(() => error.value ?? files.error.value);

  function synchronizeBlockOrder() {
    blocks.value = [...layout.value.blocks];
  }
  function beginDragging() {
    dragging.value = true;
  }
  function stopDragging() {
    dragging.value = false;
  }

  async function saveBlockOrder() {
    stopDragging();
    await editor.reorderBlocks(blocks.value.map((block) => block.id));
    synchronizeBlockOrder();
  }

  function chooseImportFile() {
    fileInput.value?.click();
  }
  function importSelectedFile(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (file) void files.importFile(file);
  }

  function addStandaloneCards() {
    editor.openAddCards(null);
  }
</script>

<template>
  <div class="nh-home" :class="{ 'is-editing': editing }">
    <UiPageHeader title="Home" :subtitle="`${areaCount} areas · ${tileCount} tiles`">
      <template #actions>
        <template v-if="editing">
          <UiButton icon="add" :disabled="busy" @click="editor.createArea">New area</UiButton>
          <UiButton icon="export" :disabled="busy" @click="files.exportLayout">Export</UiButton>
          <UiButton icon="import" :disabled="busy || files.importing.value" @click="chooseImportFile">Import</UiButton>
          <UiButton variant="primary" icon="check" :disabled="busy" @click="editor.finish">Done</UiButton>
        </template>
        <template v-else>
          <UiButton icon="panel" :to="{ name: RouteName.Panel }">Panel mode</UiButton>
          <UiButton icon="edit" :disabled="!writable || loading" @click="editor.begin">Edit layout</UiButton>
        </template>
        <input
          ref="fileInput"
          type="file"
          accept=".json,application/json"
          class="sr-only"
          tabindex="-1"
          aria-hidden="true"
          @change="importSelectedFile" />
      </template>
    </UiPageHeader>
    <div v-if="feedback" class="nh-alert is-danger" role="alert">
      {{ feedback }}
      <UiButton size="sm" :disabled="busy || loading" @click="home.load">Reload layout</UiButton>
    </div>
    <p v-if="sourceError" class="nh-alert is-warning" role="status">
      Integration results could not be updated. Device areas remain available.
    </p>
    <p v-if="editing" class="nh-alert is-info nh-home-hint" role="status">
      Drag an area by its handle, or use the arrows, to change the order. Drag a card or use Move card to place it in
      another area or directly on Home. Changes save as you go.
    </p>
    <p v-if="busy" class="sr-only" role="status">Saving Home layout</p>
    <p v-else-if="files.importing.value" class="sr-only" role="status">Importing Home layout</p>
    <div v-if="!initialized && loading && !blocks.length" aria-busy="true">Loading Home…</div>
    <UiEmpty
      v-else-if="!blocks.length"
      icon="home"
      title="No areas yet"
      text="Create an area or add individual cards to Home.">
      <UiButton :disabled="!writable || busy" @click="editor.createArea">Create your first area</UiButton>
    </UiEmpty>
    <div ref="gridRoot">
      <draggable
        v-model="blocks"
        item-key="id"
        class="nh-areas"
        :class="{ 'is-dragging': dragging }"
        handle=".nh-area-handle"
        :disabled="!editing || busy"
        :animation="160"
        ghost-class="is-ghost"
        @choose="beginDragging"
        @unchoose="stopDragging"
        @end="saveBlockOrder">
        <template #item="{ element, index }">
          <div
            class="nh-home-block"
            :class="{ 'is-area': element.kind === 'group', 'is-summary': element.cards[0]?.kind === 'waste_summary' }"
            :data-home-block="element.id">
            <div class="nh-home-block-content">
              <AreaCard
                v-if="element.kind === 'group'"
                :group="element"
                :sources="sources"
                :source-error="sourceError"
                :editing="editing"
                :busy="busy"
                :index="index"
                :count="blocks.length"
                @move="editor.moveBlock(index, $event)"
                @rename="editor.renameArea(element)"
                @remove="editor.removeArea(element)"
                @add-tiles="editor.openAddCards(element.id)"
                @remove-tile="editor.removeCard"
                @move-card="editor.openMoveCard"
                @transfer-card="editor.transferCard($event, element.id)"
                @reorder-cards="editor.reorderCards(element.id, $event)" />
              <HomeStandaloneCard
                v-else
                :block="element"
                :sources="sources"
                :source-error="sourceError"
                :editing="editing"
                :busy="busy"
                :index="index"
                :count="blocks.length"
                @remove="editor.removeCard"
                @delete-block="editor.removeStandaloneCard"
                @move-card="editor.openMoveCard"
                @move-block="editor.moveBlock(index, $event)" />
            </div>
          </div>
        </template>
      </draggable>
    </div>
    <UiButton v-if="editing" class="nh-home-add-card" icon="add" :disabled="busy" @click="addStandaloneCards">
      Add cards to Home
    </UiButton>
    <AreaNameDialog
      :open="areaDialogOpen"
      :current="areaToRename?.name"
      :busy="busy"
      :validate="editor.validateAreaName"
      @close="editor.closeArea"
      @save="editor.saveAreaName" />
    <AddTilesDialog
      :open="!!addDialog"
      :area-name="addDestination?.name ?? 'Home'"
      :existing="addDestination?.cards ?? []"
      :sources="sources"
      :source-error="sourceError"
      :busy="busy"
      @close="editor.closeAddCards"
      @add="editor.addCards" />
    <MoveCardDialog
      :card="cardToMove"
      :blocks="layout.blocks"
      :busy="busy"
      @close="editor.closeMoveCard"
      @save="editor.saveCardMove" />
  </div>
</template>

<style scoped>
  /* Disabled controls suppress Chromium mouseup; let Home drags finish over them. */
  .nh-home.is-editing :deep(button:disabled) {
    pointer-events: none;
  }
  .nh-home {
    container-type: inline-size;
    --nh-home-heading-height: 2rem;
    --nh-home-heading-gap: 0.6rem;
  }
  .nh-areas {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 9rem), 1fr));
    grid-auto-rows: 2px;
    gap: 0 var(--nh-gap);
    align-items: start;
  }
  .nh-home-block {
    min-width: 0;
    align-self: start;
    grid-row: auto / span 6;
  }
  .nh-home-block-content {
    padding-bottom: var(--nh-gap);
  }
  .nh-home:not(.is-editing) .nh-home-block.is-first-row:not(.is-area) .nh-home-block-content {
    padding-top: calc(var(--nh-home-heading-height) + var(--nh-home-heading-gap));
  }
  .nh-home :deep(.nh-home-edit-surface) {
    min-width: 0;
    border-radius: var(--nh-radius-l);
    transition: outline-color var(--nh-motion), background var(--nh-motion);
  }
  .nh-home :deep(.nh-home-edit-surface.is-editing) {
    outline: 2px dashed var(--nh-border-strong);
    outline-offset: 0.1rem;
    background: color-mix(in srgb, var(--nh-surface) 40%, transparent);
  }
  .nh-home-block.is-area,
  .nh-home-block.is-summary {
    grid-column: span 2;
  }
  .nh-home-block:not(.is-area) :deep(.nh-tiles) {
    grid-template-columns: 1fr;
  }
  .nh-home-hint {
    margin-bottom: calc(var(--nh-space) * 1.2);
  }
  .nh-home-add-card {
    margin-top: 1.5rem;
  }
  .nh-areas.is-dragging :deep(.nh-tiles) {
    display: none;
  }
  .nh-areas :deep(.is-ghost) {
    opacity: 0.4;
  }
  .nh-alert {
    margin-bottom: 1rem;
  }
  @container (max-width: 340px) {
    .nh-areas {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
</style>
