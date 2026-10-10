<script setup lang="ts">
  /*
   * Settings (spec 007 US-05). Appearance is stored in this browser; the other sections are the
   * hub settings editors kept from the previous UI (same commands), restyled by the theme.
   * The open section is kept in the URL (?section=).
   */
  import { computed, defineAsyncComponent } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { settingsTabComponents } from '@/mixins/useTabComponents';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';

  const AppearanceForm = defineAsyncComponent(() => import('@/components/shell/AppearanceForm.vue'));

  const SECTIONS = [
    { id: 'appearance', title: 'Appearance', content: AppearanceForm, note: 'Saved in this browser only.' },
    ...settingsTabComponents.map((t) => ({ id: t.title.toLowerCase().replace(/\s+/g, '-'), title: t.title, content: t.content, note: 'Saved on the hub.' })),
  ];

  const route = useRoute();
  const router = useRouter();
  const current = computed(() => SECTIONS.find((s) => s.id === route.query.section) ?? SECTIONS[0]);
  const select = (id: string) => router.replace({ query: { ...route.query, section: id === SECTIONS[0].id ? undefined : id } });
</script>

<template>
  <div>
    <UiPageHeader title="Settings" />
    <div class="nh-settings">
      <nav class="nh-vtabs" aria-label="Settings sections">
        <a
          v-for="s in SECTIONS"
          :key="s.id"
          href="#"
          class="nh-vtab"
          :aria-current="current.id === s.id ? 'page' : undefined"
          @click.prevent="select(s.id)">
          {{ s.title }}
        </a>
      </nav>
      <UiCard :title="current.title" class="nh-settings-body">
        <p class="nh-muted nh-settings-note">{{ current.note }}</p>
        <component :is="current.content" :key="current.id" />
      </UiCard>
    </div>
  </div>
</template>

<style scoped>
  .nh-settings {
    display: grid;
    grid-template-columns: 13rem minmax(0, 1fr);
    gap: calc(var(--nh-space) * 1.2);
    align-items: start;
  }
  .nh-vtabs {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  .nh-vtab {
    padding: 0.55rem 0.8rem;
    border-radius: var(--nh-radius-m);
    text-decoration: none;
    color: var(--nh-text-muted);
    font-weight: 500;
  }
  .nh-vtab:hover {
    background: var(--nh-surface-2);
    color: var(--nh-text);
  }
  .nh-vtab[aria-current='page'] {
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
    font-weight: 650;
  }
  .nh-settings-body {
    max-width: 48rem;
    min-width: 0;
  }
  .nh-settings-note {
    margin: 0 0 var(--nh-space);
    font-size: 0.85rem;
  }
  @media (max-width: 860px) {
    .nh-settings {
      grid-template-columns: minmax(0, 1fr);
    }
    .nh-vtabs {
      flex-direction: row;
      overflow-x: auto;
      scrollbar-width: none;
    }
    .nh-vtab {
      white-space: nowrap;
    }
  }
</style>
