import { computed, readonly, ref, type Ref } from 'vue';
import type { HomeBlock, HomeCard } from '@/types/home.type';
import type { HomeLayoutScope } from './useHomeLayout';
import { confirm } from './useConfirm';
import {
  addHomeCards,
  findHomePlacement,
  homeAreaNameError,
  moveHomeBlock,
  moveHomeCard,
  nameHomeArea,
  reorderHomeBlocks,
  reorderHomeCards,
  withoutHomeBlock,
  withoutHomeCard,
} from '@/domain/home-layout';

/** Coordinate Home edit dialogs and named commands. Domain helpers create proposals; the session
 * owns persistence and accepted state. Dialogs close only after an accepted save. */
export function useHomeEditor(home: Pick<HomeLayoutScope, 'layout' | 'save'>, editing: Ref<boolean>) {
  const areaDialogOpen = ref(false);
  const areaToRename = ref<HomeBlock | null>(null);
  const addDialog = ref<{ destination: string | null } | null>(null);
  const cardToMove = ref<HomeCard | null>(null);
  const addDestination = computed(() =>
    home.layout.value.blocks.find((block) => block.id === addDialog.value?.destination)
  );

  /** Enter the layout editor; refresh pausing is supplied to the page session by the caller. */
  function begin() {
    editing.value = true;
  }

  /** Leave the editor; already accepted save-as-you-go changes remain persisted. */
  function finish() {
    editing.value = false;
  }

  /** Open an empty name dialog for a new area. */
  function createArea() {
    areaToRename.value = null;
    areaDialogOpen.value = true;
  }

  /** Open a rename dialog for one existing stable area identity. */
  function renameArea(area: HomeBlock) {
    areaToRename.value = area;
    areaDialogOpen.value = true;
  }

  /** Dismiss an area dialog without creating a proposal. */
  function closeArea() {
    areaDialogOpen.value = false;
    areaToRename.value = null;
  }

  /** Validate the entered label against current accepted areas, excluding the renamed identity. */
  function validateAreaName(name: string): string | null {
    return homeAreaNameError(home.layout.value, name, areaToRename.value?.id);
  }

  /** Create or rename the selected area; a failed save retains the open dialog and input. */
  async function saveAreaName(name: string) {
    const proposal = nameHomeArea(home.layout.value, name, areaToRename.value?.id);
    if (await home.save(proposal)) closeArea();
  }

  /** Confirm removal of an area's placements, preserving its devices/integrations at their sources. */
  async function removeArea(area: HomeBlock) {
    const accepted = await confirm({
      title: `Delete ${area.name}?`,
      message:
        'The area and its card placements are removed from Home and panel. Devices, integrations and saved results are kept.',
      confirmLabel: 'Delete area',
      danger: true,
    });
    if (accepted) await home.save(withoutHomeBlock(home.layout.value, area.id));
  }

  /** Open the picker for one area, or null for standalone cards on Home. */
  function openAddCards(destination: string | null) {
    addDialog.value = { destination };
  }

  /** Dismiss the picker without changing accepted membership. */
  function closeAddCards() {
    addDialog.value = null;
  }

  /** Place selected source references at the selected destination, retaining the picker on failure. */
  async function addCards(cards: HomeCard[]) {
    if (!addDialog.value || cards.length === 0) return;
    const proposal = addHomeCards(home.layout.value, addDialog.value.destination, cards);
    if (await home.save(proposal)) closeAddCards();
  }

  /** Open placement/panel preferences for one card. */
  function openMoveCard(card: HomeCard) {
    cardToMove.value = card;
  }

  /** Dismiss placement preferences without changing membership. */
  function closeMoveCard() {
    cardToMove.value = null;
  }

  /** Move the selected card and update its panel visibility in one accepted transaction. */
  async function saveCardMove(destination: string | null, panel: boolean) {
    const card = cardToMove.value;
    if (!card) return;
    const proposal = moveHomeCard(home.layout.value, card.id, destination, panel);
    if (await home.save(proposal)) closeMoveCard();
  }

  /** Accept a drag transfer at its actual destination index; never save a transient source removal. */
  async function transferCard(placement: { id: string; index: number }, destination: string) {
    const accepted = findHomePlacement(home.layout.value.blocks, placement.id);
    if (!accepted) return;
    const { card } = accepted;
    await home.save(moveHomeCard(home.layout.value, card.id, destination, card.panel, placement.index));
  }

  /** Remove a placement; its source configuration/result is unaffected. */
  async function removeCard(id: string) {
    await home.save(withoutHomeCard(home.layout.value, id));
  }

  /** Confirm removal of a standalone placement without changing its source. */
  async function removeStandaloneCard(card: HomeCard, title: string) {
    const accepted = await confirm({
      title: `Delete ${title}?`,
      message: 'This card is removed from Home and panel. Its source and saved results are kept.',
      confirmLabel: 'Delete card',
      danger: true,
    });
    if (accepted) await removeCard(card.id);
  }

  /** Move a block by accepted reading-order positions. */
  async function moveBlock(from: number, to: number) {
    await home.save(moveHomeBlock(home.layout.value, from, to));
  }

  /** Persist an owned drag order of block IDs. */
  async function reorderBlocks(ids: readonly string[]) {
    await home.save(reorderHomeBlocks(home.layout.value, ids));
  }

  /** Persist only the existing card IDs in an area's new order. */
  async function reorderCards(groupId: string, cards: readonly HomeCard[]) {
    await home.save(
      reorderHomeCards(
        home.layout.value,
        groupId,
        cards.map((card) => card.id)
      )
    );
  }

  return {
    areaDialogOpen: readonly(areaDialogOpen),
    areaToRename: readonly(areaToRename),
    addDialog: readonly(addDialog),
    addDestination,
    cardToMove: readonly(cardToMove),
    begin,
    finish,
    createArea,
    renameArea,
    closeArea,
    validateAreaName,
    saveAreaName,
    removeArea,
    openAddCards,
    closeAddCards,
    addCards,
    openMoveCard,
    closeMoveCard,
    saveCardMove,
    transferCard,
    removeCard,
    removeStandaloneCard,
    moveBlock,
    reorderBlocks,
    reorderCards,
  };
}
