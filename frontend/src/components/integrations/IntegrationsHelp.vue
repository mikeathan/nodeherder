<script setup lang="ts">
  import { computed, ref, onMounted, nextTick } from 'vue';
  import { useRoute } from 'vue-router';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import { helpTopics } from '@/domain/integrations-help';
  import { RouteName } from '@/types/router';
  import './connections.css';
  const route = useRoute();
  const search = ref('');
  const topics = computed(() => helpTopics(search.value));
  onMounted(async () => {
    await nextTick();
    if (route.hash) {
      const topic = document.getElementById(route.hash.slice(1));
      topic?.focus();
      topic?.scrollIntoView({ block: 'start' });
    }
  });
</script>
<template>
  <div class="connections connection-stack">
    <UiPageHeader
      title="Connections help"
      :back="{ name: RouteName.ConnectionProvider, params: route.params, query: route.query }"
      back-label="Configure operations" /><label class="connection-field"
      ><span>Find help</span
      ><input v-model="search" type="search" placeholder="For example: change URL or collection fields"
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
