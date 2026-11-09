<script setup lang="ts">
  /*
   * Settings host (spec 007 US-05, spec 059 FR-01). Sections come from the registry so ids
   * never depend on display titles. The open section is kept in the URL (?section=) and
   * the first section keeps a clean URL.
   */
  import { computed } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import {
    SETTINGS_SECTIONS,
    groupSettingsSections,
    resolveSettingsSection,
    sectionQueryValue,
  } from './sections';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const route = useRoute();
  const router = useRouter();

  const groups = groupSettingsSections();
  const current = computed(() => resolveSettingsSection(route.query.section));

  function select(id: string) {
    const section = SETTINGS_SECTIONS.find((candidate) => candidate.id === id);
    if (!section || section.id === current.value.id) return;
    router.replace({ query: { ...route.query, section: sectionQueryValue(section) } });
  }
</script>

<template>
  <div>
    <UiPageHeader title="Settings" subtitle="Hub behaviour for this home, plus appearance for this browser." />
    <div class="nh-settings">
      <nav class="nh-vtabs" aria-label="Settings sections">
        <div v-for="group in groups" :key="group.group" class="nh-vgroup">
          <p :id="`settings-group-${group.group}`" class="nh-vgroup-label">{{ group.label }}</p>
          <ul :aria-labelledby="`settings-group-${group.group}`">
            <li v-for="section in group.sections" :key="section.id">
              <a
                href="#"
                class="nh-vtab"
                :aria-current="current.id === section.id ? 'page' : undefined"
                @click.prevent="select(section.id)">
                <UiIcon :name="section.icon" />
                {{ section.title }}
              </a>
            </li>
          </ul>
        </div>
      </nav>

      <div class="nh-settings-body">
        <component :is="current.content" :key="current.id" />
      </div>
    </div>
  </div>
</template>

<style scoped>
  .nh-settings {
    display: grid;
    grid-template-columns: 14rem minmax(0, 1fr);
    gap: calc(var(--nh-space) * 1.2);
    align-items: start;
  }
  .nh-settings-body {
    max-width: 56rem;
    min-width: 0;
  }
  .nh-vtabs {
    position: sticky;
    top: calc(var(--nh-header-h) + var(--nh-space));
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .nh-vgroup ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  .nh-vgroup-label {
    margin: 0 0 0.35rem;
    padding: 0 0.8rem;
    font-size: 0.72rem;
    font-weight: 650;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--nh-text-muted);
  }
  .nh-vtab {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0.55rem 0.8rem;
    border-radius: var(--nh-radius-m);
    text-decoration: none;
    color: var(--nh-text-muted);
    font-weight: 500;
  }
  .nh-vtab .nh-ic {
    width: 1.05rem;
    height: 1.05rem;
    opacity: 0.85;
  }
  .nh-vtab[aria-current='page'] .nh-ic {
    opacity: 1;
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
  @media (max-width: 860px) {
    .nh-settings {
      grid-template-columns: minmax(0, 1fr);
    }
    .nh-vtabs {
      position: static;
      flex-direction: row;
      align-items: center;
      gap: 0.75rem;
      overflow-x: auto;
      padding-bottom: 0.35rem;
      scrollbar-width: none;
    }
    .nh-vgroup {
      display: flex;
      align-items: center;
      gap: 0.4rem;
    }
    .nh-vgroup-label {
      margin: 0;
      padding: 0;
      white-space: nowrap;
    }
    .nh-vgroup ul {
      flex-direction: row;
      gap: 0.15rem;
    }
    .nh-vtab {
      white-space: nowrap;
    }
  }
</style>
