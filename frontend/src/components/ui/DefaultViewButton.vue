<script setup lang="ts">
  import { computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { useDefaultView } from '@/composables/useDefaultView';
  import { isDefaultView } from '@/domain/default-view';
  import UiIcon from './UiIcon.vue';

  withDefaults(defineProps<{ dock?: boolean }>(), { dock: false });
  const route = useRoute();
  const { preferred, toggle } = useDefaultView();
  const eligible = computed(() => isDefaultView(route.name));
  const pinned = computed(() => preferred.value === route.name);
  const hint = computed(() => pinned.value ? 'Unpin view as the default on this browser' : 'Pin view as the default on this browser');

  function toggleCurrent() {
    if (isDefaultView(route.name)) toggle(route.name);
  }
</script>

<template>
  <button
    v-if="eligible"
    type="button"
    class="nh-btn is-ghost nh-default-view"
    :class="{ 'is-pinned': pinned, 'is-dock': dock }"
    :aria-pressed="pinned"
    :aria-label="hint"
    :title="hint"
    @click="toggleCurrent">
    <UiIcon :name="pinned ? 'pin' : 'pinOutline'" />
    <span>{{ pinned ? 'Pinned' : 'Pin view' }}</span>
  </button>
</template>

<style scoped>
  .nh-default-view {
    flex: none;
    font-size: 0.8rem;
  }
  .nh-default-view.is-pinned {
    color: var(--nh-accent);
  }
  @media (max-width: 760px) {
    .nh-default-view:not(.is-dock) {
      width: 2.5rem;
      padding: 0;
    }
    .nh-default-view:not(.is-dock) span {
      display: none;
    }
  }
</style>
