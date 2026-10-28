<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { findHomePlacement } from '@/domain/home-layout';
  import type { HomeBlock, HomeCard } from '@/types/home.type';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';

  const props = defineProps<{ card: HomeCard | null; blocks: readonly HomeBlock[]; busy: boolean }>();
  const emit = defineEmits<{ close: []; save: [destination: string | null, panel: boolean] }>();
  const destination = ref<string | null>(null);
  const panel = ref(true);
  const choices = computed(() => [
    { value: null, label: 'Directly on Home' },
    ...props.blocks.filter((block) => block.kind === 'group').map((block) => ({ value: block.id, label: block.name })),
  ]);
  watch(
    () => props.card,
    (card) => {
      const placement = card ? findHomePlacement(props.blocks, card.id) : null;
      const owner = placement?.block;
      destination.value = owner?.kind === 'group' ? owner.id : null;
      panel.value = card?.panel ?? true;
    }
  );
</script>

<template>
  <UiDialog :open="!!card" title="Move card" @close="emit('close')">
    <form id="move-home-card" @submit.prevent="emit('save', destination, panel)">
      <label class="nh-label" for="home-card-destination">Place card</label>
      <UiSelect id="home-card-destination" v-model="destination" :options="choices" :disabled="busy" />
      <label class="nh-home-panel-check">
        <input v-model="panel" type="checkbox" :disabled="busy" />
        Show this card in panel mode
      </label>
      <p class="nh-muted">
        Moving a card keeps its source and current result. Panel visibility does not change its Home position.
      </p>
    </form>
    <template #footer>
      <UiButton :disabled="busy" @click="emit('close')">Cancel</UiButton>
      <UiButton variant="primary" type="submit" form="move-home-card" :disabled="busy">Save placement</UiButton>
    </template>
  </UiDialog>
</template>

<style scoped>
  .nh-home-panel-check {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin: 1.2rem 0;
    min-height: 2.75rem;
  }
  .nh-home-panel-check input {
    accent-color: var(--nh-accent);
    width: 1.1rem;
    height: 1.1rem;
  }
</style>
