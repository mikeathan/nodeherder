<script setup lang="ts">
  import { computed, ref, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import IntegrationPageHeader from '../shared/IntegrationPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import { COMMON_HELP, matchesHelpTopic, type HelpTopic } from '@/domain/integrations/help';
  import { integrationHelp } from '@/integrations/registry';
  import IntegrationSections from '../shared/IntegrationSections.vue';
  import { useProviderSections } from '@/composables/integrations/useProviderSections';
  import '../connections.css';
  const route = useRoute();
  const sectionContext = useProviderSections();
  const search = ref('');
  const typeId = computed(() => (typeof route.params.type === 'string' ? route.params.type : ''));
  const providerId = computed(() => (typeof route.params.provider === 'string' ? route.params.provider : ''));
  const topics = computed<HelpTopic[]>(() =>
    [...integrationHelp(typeId.value, providerId.value), ...COMMON_HELP].filter((topic) =>
      matchesHelpTopic(topic, search.value)
    )
  );
  onMounted(() => {
    if (route.hash) {
      const topic = document.getElementById(route.hash.slice(1));
      topic?.focus();
      topic?.scrollIntoView({ block: 'start' });
    }
  });
</script>
<template>
  <div class="connections connection-stack">
    <IntegrationPageHeader
      title="Integrations help"
      subtitle="Shared guidance plus provider-specific setup notes." />
    <IntegrationSections active="help" :context="sectionContext" />
    <label class="connection-field"
      ><span>Find help</span
      ><input v-model="search" type="search" placeholder="For example: change URL or field mapping"
    /></label>
    <div class="connection-help-layout">
      <nav class="connection-stack" aria-label="Help topics">
        <a v-for="topic in topics" :key="topic.id" :href="`#${topic.id}`">{{ topic.title }}</a>
      </nav>
      <div class="connection-stack">
        <UiCard
          v-for="topic in topics"
          :id="topic.id"
          :key="topic.id"
          tabindex="-1"
          :title="topic.title"
          class="connection-help-topic"
          ><p v-for="paragraph in topic.paragraphs" :key="paragraph">{{ paragraph }}</p></UiCard
        >
        <p v-if="!topics.length" role="status">No matching topics. Try a shorter search.</p>
      </div>
    </div>
  </div>
</template>
