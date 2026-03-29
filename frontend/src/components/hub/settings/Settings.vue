<script setup lang="ts">
  import { ref } from 'vue';
  import { settingsTabComponents as settingsTabs } from '../../../mixins/useTabComponents';

  const activeTab = ref('0');
</script>

<template>
  <Card>
    <template #title>
      <div class="text-3xl align-self-center">Settings</div>
    </template>
    <template #content>
      <div class="col-12 md:col-12">
        <Tabs v-model:value="activeTab">
          <TabList>
            <Tab
              v-for="tab in settingsTabs"
              :key="tab.value"
              :value="tab.value">
              {{ tab.title }}
            </Tab>
          </TabList>
          <TabPanels>
            <TabPanel v-for="tab in settingsTabs" :key="tab.value" :value="tab.value">
              <component v-if="activeTab === tab.value" :is="tab.content"></component>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </template>
  </Card>
</template>
